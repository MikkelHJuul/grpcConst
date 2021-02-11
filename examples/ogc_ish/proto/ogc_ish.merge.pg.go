package ogs_ish

func (receiver *Feature) DoMerge(donor *Feature) {
	if donor == nil {
		return
	}
	if receiver.Type == "" {
		receiver.Type = donor.Type
	}
	if receiver.Id == "" {
		receiver.Id = donor.Id
	}
	if receiver.Properties == nil {
		receiver.Properties = donor.Properties
	} else if donor.Properties != nil {
		properties := receiver.Properties
		dProps := donor.Properties
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
		receiver.Geometry = donor.Geometry
	} else if donor.Geometry != nil {
		geom := receiver.Geometry
		dGeom := donor.Geometry
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
}
