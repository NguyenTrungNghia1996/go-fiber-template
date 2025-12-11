package controllers

import (
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
