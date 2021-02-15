package ogs_ish

func (receiver *Feature) MergeFieldsFrom(donor interface{}) {
	if d, ok := donor.(Feature); ok {
		if receiver.Type == "" {
			receiver.Type = d.Type
		}
		if receiver.Id == "" {
			receiver.Id = d.Id
		}
		if receiver.Properties == nil {
			receiver.Properties = d.Properties
		} else if d.Properties != nil {
			properties := receiver.Properties
			dProps := d.Properties
			if properties.Measurement == nil {
				properties.Measurement = dProps.Measurement
			} else if dProps.Measurement != nil {
				if properties.Measurement.Value == 0 {
					properties.Measurement.Value = dProps.Measurement.Value
				}
				if properties.Measurement.Name == "" {
					properties.Measurement.Name = dProps.Measurement.Name
				}
			}
			if properties.Station == nil {
				properties.Station = dProps.Station
			} else if dProps.Station != nil {
				if properties.Station.Metadata == "" {
					properties.Station.Metadata = dProps.Station.Metadata
				}
				if properties.Station.Name == "" {
					properties.Station.Name = dProps.Station.Name
				}
			}
		}
		if receiver.Geometry == nil {
			receiver.Geometry = d.Geometry
		} else if d.Geometry != nil {
			geom := receiver.Geometry
			dGeom := d.Geometry
			if geom.Type == "" {
				geom.Type = dGeom.Type
			}
			if geom.Coordinates == nil {
				geom.Coordinates = dGeom.Coordinates
			} else if dGeom.Coordinates != nil {
				if geom.Coordinates.Longitude == 0 {
					geom.Coordinates.Longitude = dGeom.Coordinates.Longitude
				}
				if geom.Coordinates.Latitude == 0 {
					geom.Coordinates.Latitude = dGeom.Coordinates.Latitude
				}
			}
		}
		if receiver.unknownFields != nil {
			receiver.unknownFields = d.unknownFields
		}
	}
}
