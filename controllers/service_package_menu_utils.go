package controllers

import (
	"errors"
	"strings"

	"go-fiber-api/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ensureMenuIDs assigns an ObjectID to any menu that is missing one.
// It returns true when at least one ID was added.
func ensureMenuIDs(menus []models.ServicePackageMenu) bool {
	changed := false
	for i := range menus {
		if menus[i].ID.IsZero() {
			menus[i].ID = primitive.NewObjectID()
			changed = true
		}
	}
	return changed
}

// trimMenuFields trims string fields of a menu in place.
func trimMenuFields(menu *models.ServicePackageMenu) {
	menu.Title = strings.TrimSpace(menu.Title)
	menu.Key = strings.TrimSpace(menu.Key)
	menu.URL = strings.TrimSpace(menu.URL)
	menu.Icon = strings.TrimSpace(menu.Icon)
}

func parseServicePackageParentID(raw string) (models.ObjectIDOrNil, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return models.ObjectIDOrNil(primitive.NilObjectID), nil
	}
	oid, err := primitive.ObjectIDFromHex(raw)
	if err != nil {
		return models.ObjectIDOrNil(primitive.NilObjectID), errors.New("parent_id must be a valid ObjectID")
	}
	return models.ObjectIDOrNil(oid), nil
}

func buildServicePackageMenus(inputs []models.ServicePackageMenuPayload) ([]models.ServicePackageMenu, error) {
	menus := make([]models.ServicePackageMenu, len(inputs))
	for i := range inputs {
		parentID, err := parseServicePackageParentID(inputs[i].ParentID)
		if err != nil {
			return nil, err
		}
		menu := models.ServicePackageMenu{
			Title:      inputs[i].Title,
			Key:        inputs[i].Key,
			URL:        inputs[i].URL,
			Icon:       inputs[i].Icon,
			ParentID:   parentID,
			Permission: inputs[i].Permission,
			Active:     inputs[i].Active,
		}
		if trimmed := strings.TrimSpace(inputs[i].ID); trimmed != "" {
			oid, err := primitive.ObjectIDFromHex(trimmed)
			if err != nil {
				return nil, errors.New("menu id must be a valid ObjectID")
			}
			menu.ID = oid
		}
		trimMenuFields(&menu)
		menus[i] = menu
	}
	ensureMenuIDs(menus)
	return menus, nil
}
