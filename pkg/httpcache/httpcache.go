package httpcache

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	fibercache "github.com/gofiber/fiber/v2/middleware/cache"
	memorystorage "github.com/gofiber/storage/memory"
)

const defaultTTL = 30 * time.Second

// Config holds cache settings loaded from environment variables.
type Config struct {
	Enabled    bool
	Expiration time.Duration
	MaxBytes   uint
}

var (
	initOnce sync.Once
	handler  fiber.Handler
	store    *memorystorage.Storage
	cfg      Config
)

// Middleware returns a shared caching middleware for GET/HEAD requests.
func Middleware() fiber.Handler {
	initOnce.Do(initCache)
	return handler
}

// InvalidateOnWrite clears cached GET responses after successful write operations.
func InvalidateOnWrite() fiber.Handler {
	initOnce.Do(initCache)
	if !cfg.Enabled {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return func(c *fiber.Ctx) error {
		method := c.Method()
		if method != fiber.MethodPost && method != fiber.MethodPut && method != fiber.MethodDelete && method != fiber.MethodPatch {
			return c.Next()
		}
		if strings.HasPrefix(c.Path(), "/auth/") {
			return c.Next()
		}
		if err := c.Next(); err != nil {
			return err
		}
		status := c.Response().StatusCode()
		if status == 0 || status < fiber.StatusBadRequest {
			clearCache()
		}
		return nil
	}
}

// GetConfig exposes the loaded cache config for logging.
func GetConfig() Config {
	initOnce.Do(initCache)
	return cfg
}

// Clear forces a cache reset.
func Clear() {
	initOnce.Do(initCache)
	clearCache()
}

func initCache() {
	cfg = loadConfig()
	if !cfg.Enabled {
		handler = func(c *fiber.Ctx) error {
			return c.Next()
		}
		return
	}

	store = memorystorage.New()
	cacheHandler := fibercache.New(fibercache.Config{
		Expiration:           cfg.Expiration,
		Storage:              store,
		KeyGenerator:         cacheKey,
		StoreResponseHeaders: true,
		MaxBytes:             cfg.MaxBytes,
	})

	handler = func(c *fiber.Ctx) error {
		if c.Method() != fiber.MethodGet && c.Method() != fiber.MethodHead {
			return c.Next()
		}
		return cacheHandler(c)
	}
}

func cacheKey(c *fiber.Ctx) string {
	uri := c.OriginalURL()
	if uri == "" {
		uri = c.Path()
	}
	var b strings.Builder
	b.Grow(len(uri) + 80)
	b.WriteString(uri)
	if auth := strings.TrimSpace(c.Get("Authorization")); auth != "" {
		sum := sha256.Sum256([]byte(auth))
		b.WriteString("|auth=")
		b.WriteString(hex.EncodeToString(sum[:]))
	}
	return b.String()
}

func loadConfig() Config {
	enabled := true
	if raw := strings.TrimSpace(os.Getenv("CACHE_ENABLED")); raw != "" {
		enabled = parseBool(raw)
	}

	ttl := defaultTTL
	if raw := strings.TrimSpace(os.Getenv("CACHE_TTL_SECONDS")); raw != "" {
		if sec, err := strconv.Atoi(raw); err == nil && sec > 0 {
			ttl = time.Duration(sec) * time.Second
		}
	}

	var maxBytes uint
	if raw := strings.TrimSpace(os.Getenv("CACHE_MAX_BYTES")); raw != "" {
		if v, err := strconv.ParseUint(raw, 10, 64); err == nil {
			maxBytes = uint(v)
		}
	}

	return Config{
		Enabled:    enabled,
		Expiration: ttl,
		MaxBytes:   maxBytes,
	}
}

func parseBool(s string) bool {
	switch strings.ToLower(s) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return false
	}
}

func clearCache() {
	if store != nil {
		_ = store.Reset()
	}
}
