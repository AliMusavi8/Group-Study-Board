package models

import "time"

type Point struct {
	X float64 `json:"x" bson:"x"`
	Y float64 `json:"y" bson:"y"`
}

type ClientEvent struct {
	Type      string  `json:"type"`
	RoomID    string  `json:"roomId,omitempty"`
	ClientID  string  `json:"clientId,omitempty"`
	Point     *Point  `json:"point,omitempty"`
	Color     string  `json:"color,omitempty"`
	Thickness float64 `json:"thickness,omitempty"`
	Tool      string  `json:"tool,omitempty"`
}

type Event struct {
	RoomID    string  `json:"roomId" bson:"roomId"`
	Type      string  `json:"type" bson:"type"`
	ClientID  string  `json:"clientId" bson:"clientId"`
	Point     *Point  `json:"point,omitempty" bson:"point,omitempty"`
	Color     string  `json:"color,omitempty" bson:"color,omitempty"`
	Thickness float64 `json:"thickness,omitempty" bson:"thickness,omitempty"`
	Tool      string  `json:"tool,omitempty" bson:"tool,omitempty"`
	Seq       int64   `json:"seq" bson:"seq"`
	ServerTs  int64   `json:"serverTs" bson:"serverTs"`
	CreatedAt time.Time `json:"-" bson:"createdAt"`
}

type SnapshotPayload struct {
	Events    []Event `json:"events" bson:"events"`
	CreatedAt int64   `json:"createdAt" bson:"createdAt"`
}

type SnapshotMessage struct {
	Type     string           `json:"type"`
	RoomID   string           `json:"roomId"`
	Snapshot *SnapshotPayload `json:"snapshot,omitempty"`
	Events   []Event          `json:"events"`
}

type EventMessage struct {
	Type  string `json:"type"`
	Event Event  `json:"event"`
}

type ErrorMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
