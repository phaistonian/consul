package dataplane

import (
	"context"

	"github.com/hashicorp/consul/proto-public/pbdataplane"
)

func (d *Server) SupportedDataplaneFeatures(ctx context.Context, req *pbdataplane.SupportedDataplaneFeaturesRequest) (*pbdataplane.SupportedDataplaneFeaturesResponse, error) {
	// authz, err := d.Backend.ResolveToken(req.Token)
	// if err != nil {
	// 	return err
	// }

	supportedFeatures := []*pbdataplane.DataplaneFeatureSupport{
		{
			FeatureName: pbdataplane.DataplaneFeatures_WATCH_SERVERS,
			Supported:   true,
		},
		{
			FeatureName: pbdataplane.DataplaneFeatures_EDGE_CERTIFICATE_MANAGEMENT,
			Supported:   true,
		},
		{
			FeatureName: pbdataplane.DataplaneFeatures_ENVOY_BOOTSTRAP_CONFIGURATION,
			Supported:   true,
		},
	}
	return &pbdataplane.SupportedDataplaneFeaturesResponse{SupportedDataplaneFeatures: supportedFeatures}, nil
}
