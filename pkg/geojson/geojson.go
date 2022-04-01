package geojson

type GeoJSON struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
}
