package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// Organization represents a tenant or company entity.
type Organization struct {
    ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Name        string             `json:"name" bson:"name"`
    Description string             `json:"description" bson:"description"`
}

// OrganizationResponse is used for detailed responses.
type OrganizationResponse struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
}

// OrganizationListItem is used in list results.
type OrganizationListItem struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
}

func (o Organization) ToResponse() OrganizationResponse {
    return OrganizationResponse{
        ID:          o.ID.Hex(),
        Name:        o.Name,
        Description: o.Description,
    }
}

func (o Organization) ToListItem() OrganizationListItem {
    return OrganizationListItem{
        ID:          o.ID.Hex(),
        Name:        o.Name,
        Description: o.Description,
    }
}

