package controllers

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-fiber-api/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type UploadController struct {
	client    *minio.Client
	region    string
	endpoint  string
	accessKey string
	secretKey string
	useSSL    bool
	bucket    string
	publicURL string
}

// NewUploadController reads env and stores config. Client will be created lazily.
func NewUploadController() *UploadController {
	rawEndpoint := strings.TrimSpace(os.Getenv("MINIO_ENDPOINT"))
	useSSL := strings.EqualFold(strings.TrimSpace(os.Getenv("MINIO_SSL")), "true")
	// allow MINIO_ENDPOINT with scheme; prefer scheme over MINIO_SSL if provided
	if strings.HasPrefix(rawEndpoint, "http://") || strings.HasPrefix(rawEndpoint, "https://") {
		if u, err := url.Parse(rawEndpoint); err == nil {
			if u.Scheme == "http" || u.Scheme == "https" {
				useSSL = u.Scheme == "https"
			}
			rawEndpoint = u.Host
		}
	}
	return &UploadController{
		endpoint:  rawEndpoint,
		region:    strings.TrimSpace(os.Getenv("MINIO_REGION")),
		accessKey: strings.TrimSpace(os.Getenv("MINIO_ACCESS_KEY")),
		secretKey: strings.TrimSpace(os.Getenv("MINIO_SECRET_KEY")),
		bucket:    strings.TrimSpace(os.Getenv("MINIO_BUCKET")),
		useSSL:    useSSL,
		publicURL: strings.TrimSpace(os.Getenv("MINIO_PUBLIC_URL")),
	}
}

func (h *UploadController) ensureClient() error {
	if h.client != nil {
		return nil
	}
	if h.endpoint == "" || h.accessKey == "" || h.secretKey == "" || h.bucket == "" {
		return errors.New("upload not configured")
	}
	opt := &minio.Options{
		Creds:  credentials.NewStaticV4(h.accessKey, h.secretKey, ""),
		Secure: h.useSSL,
		Region: h.region,
	}
	c, err := minio.New(h.endpoint, opt)
	if err != nil {
		return err
	}
	h.client = c
	return nil
}

type presignInput struct {
	Filename string `json:"filename"`
	Filetype string `json:"filetype"`
}

// PresignedURL handles PUT /uploads/presigned_url (admin)
// or PUT /unit_uploads/presigned_url (unit user) and returns a pre-signed PUT URL.
func (h *UploadController) PresignedURL(c *fiber.Ctx) error {
	var in presignInput
	if err := c.BodyParser(&in); err != nil {
		return response.Error(c, "invalid body", fiber.StatusBadRequest, nil)
	}
	name := strings.TrimSpace(in.Filename)
	ctype := strings.TrimSpace(in.Filetype)
	if name == "" || ctype == "" {
		return response.Error(c, "filename and filetype are required", fiber.StatusBadRequest, nil)
	}

	if err := h.ensureClient(); err != nil {
		if strings.Contains(err.Error(), "not configured") {
			return response.Error(c, "upload not configured", fiber.StatusInternalServerError, nil)
		}
		return response.Error(c, "storage error", fiber.StatusInternalServerError, nil)
	}

	// sanitize filename and keep extension
	base := filepath.Base(name)
	ext := filepath.Ext(base)
	if len(ext) > 16 { // guard against abusive long ext
		ext = ""
	}
	// Build a safe slug from the original filename (without extension)
	baseNoExt := strings.TrimSuffix(base, ext)
	slug := sanitizeFilename(baseNoExt)
	if slug == "" {
		slug = "file"
	}
	// keep slug short to avoid path length issues
	if len(slug) > 60 {
		slug = slug[:60]
	}
	id := uuid.New().String()
	fileWithName := fmt.Sprintf("%s-%s%s", id, slug, ext)

	// Namespace object key by tenant/admin:
	// - Unit user (RequireUser):  <unit_id>/<uuid-slug.ext>
	// - Super admin (RequireAdmin): admin/<uuid-slug.ext>
	unitID, _ := c.Locals("unit_id").(string)
	unitID = strings.TrimSpace(unitID)
	objectName := fileWithName
	if unitID != "" {
		objectName = fmt.Sprintf("%s/%s", unitID, fileWithName)
	} else {
		objectName = fmt.Sprintf("admin/%s", fileWithName)
	}

	// Verify bucket exists
	exists, err := h.client.BucketExists(c.Context(), h.bucket)
	if err == nil {
		if !exists {
			// Try to create the bucket if it does not exist (requires permission). If it fails, bubble up.
			createOpts := minio.MakeBucketOptions{Region: h.region}
			if err := h.client.MakeBucket(c.Context(), h.bucket, createOpts); err != nil {
				return response.Error(c, "bucket not found", fiber.StatusInternalServerError, fiber.Map{"reason": err.Error()})
			}
		}
	}

	// Generate presigned PUT URL (15 minutes)
	expiry := 15 * time.Minute
	u, err := h.client.PresignedPutObject(c.Context(), h.bucket, objectName, expiry)
	if err != nil {
		return response.Error(c, "failed to generate url", fiber.StatusInternalServerError, fiber.Map{"reason": err.Error()})
	}

	// Do not mutate the presigned URL query after signing; client can set Content-Type header.

	// Compute public URL if configured
	var public string
	if h.publicURL != "" {
		// Ensure no double slashes
		base, _ := url.Parse(h.publicURL)
		base.Path = strings.TrimRight(base.Path, "/") + "/" + objectName
		public = base.String()
	}

	return response.Success(c, fiber.Map{
		"url":        u.String(),
		"method":     "PUT",
		"headers":    fiber.Map{"Content-Type": ctype},
		"expires_in": int(expiry.Seconds()),
		"key":        objectName,
		"public_url": public,
	}, "ok")
}

// Delete handles DELETE /uploads/file?id=<key> (admin)
// or DELETE /unit_uploads/file?id=<key> (unit user) to remove an object from the bucket root.
func (h *UploadController) Delete(c *fiber.Ctx) error {
	key := strings.TrimSpace(c.Query("id"))
	if key == "" {
		return response.Error(c, "id query param is required", fiber.StatusBadRequest, nil)
	}
	if err := h.ensureClient(); err != nil {
		return response.Error(c, "storage error", fiber.StatusInternalServerError, nil)
	}
	// Check if object exists to honor 404 semantics
	if _, err := h.client.StatObject(c.Context(), h.bucket, key, minio.StatObjectOptions{}); err != nil {
		er := minio.ToErrorResponse(err)
		if er.Code == "NoSuchKey" || strings.EqualFold(er.Code, "NotFound") || er.StatusCode == 404 {
			return response.Error(c, "file not found", fiber.StatusNotFound, nil)
		}
		return response.Error(c, "failed to stat file", fiber.StatusInternalServerError, fiber.Map{"reason": err.Error()})
	}
	// Delete
	if err := h.client.RemoveObject(c.Context(), h.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return response.Error(c, "failed to delete file", fiber.StatusInternalServerError, fiber.Map{"reason": err.Error()})
	}
	return response.Success(c, true, "deleted")
}

// sanitizeFilename returns a lowercase, hyphenated version keeping [a-z0-9-_].
// Non-matching runes are replaced with '-'; repeated '-' are collapsed.
func sanitizeFilename(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	b := make([]rune, 0, len(s))
	lastHyphen := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b = append(b, r)
			lastHyphen = false
			continue
		}
		if r == '-' || r == '_' {
			if r == '-' {
				if lastHyphen {
					continue
				}
				lastHyphen = true
			} else {
				lastHyphen = false
			}
			b = append(b, r)
			continue
		}
		// replace others with '-'
		if !lastHyphen {
			b = append(b, '-')
			lastHyphen = true
		}
	}
	// trim leading/trailing '-'
	res := strings.Trim(string(b), "-")
	return res
}
