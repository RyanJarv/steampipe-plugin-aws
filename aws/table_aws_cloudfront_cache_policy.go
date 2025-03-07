package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"
)

//// TABLE DEFINITION

func tableAwsCloudFrontCachePolicy(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "aws_cloudfront_cache_policy",
		Description: "AWS CloudFront Cache Policy",
		Get: &plugin.GetConfig{
			KeyColumns:        plugin.SingleColumn("id"),
			ShouldIgnoreError: isNotFoundError([]string{"NoSuchCachePolicy"}),
			Hydrate:           getCloudFrontCachePolicy,
		},
		List: &plugin.ListConfig{
			Hydrate: listCloudFrontCachePolicies,
		},
		Columns: awsColumns([]*plugin.Column{
			{
				Name:        "name",
				Description: "A unique name to identify the cache policy.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("CachePolicy.CachePolicyConfig.Name"),
			},
			{
				Name:        "id",
				Description: "The unique identifier for the cache policy.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("CachePolicy.Id"),
			},
			{
				Name:        "comment",
				Description: "A comment to describe the cache policy.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("CachePolicy.CachePolicyConfig.Comment"),
			},
			{
				Name:        "default_ttl",
				Description: "The default amount of time, in seconds, that you want objects to stay in the CloudFront cache before CloudFront sends another request to the origin to see if the object has been updated.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("CachePolicy.CachePolicyConfig.DefaultTTL"),
			},
			{
				Name:        "etag",
				Description: "The current version of the cache policy.",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getCloudFrontCachePolicy,
				Transform:   transform.FromField("ETag"),
			},
			{
				Name:        "max_ttl",
				Description: "The maximum amount of time, in seconds, that you want objects to stay in the CloudFront cache before CloudFront sends another request to the origin to see if the object has been updated.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("CachePolicy.CachePolicyConfig.MaxTTL"),
			},
			{
				Name:        "min_ttl",
				Description: "The minimum amount of time, in seconds, that you want objects to stay in the CloudFront cache before CloudFront sends another request to the origin to see if the object has been updated.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("CachePolicy.CachePolicyConfig.MinTTL"),
			},
			{
				Name:        "last_modified_time",
				Description: "The date and time when the cache policy was last modified.",
				Type:        proto.ColumnType_TIMESTAMP,
				Transform:   transform.FromField("CachePolicy.LastModifiedTime"),
			},
			{
				Name:        "parameters_in_cache_key_and_forwarded_to_origin",
				Description: "The HTTP headers, cookies, and URL query strings to include in the cache key. The values included in the cache key are automatically included in requests that CloudFront sends to the origin.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("CachePolicy.CachePolicyConfig.ParametersInCacheKeyAndForwardedToOrigin"),
			},

			// Steampipe standard columns
			{
				Name:        "title",
				Description: resourceInterfaceDescription("title"),
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("CachePolicy.CachePolicyConfig.Name"),
			},
			{
				Name:        "akas",
				Description: resourceInterfaceDescription("akas"),
				Type:        proto.ColumnType_JSON,
				Hydrate:     getCloudfrontCachePolicyAkas,
				Transform:   transform.FromValue(),
			},
		}),
	}
}

//// LIST FUNCTION

func listCloudFrontCachePolicies(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("listCloudFrontCachePolicies")

	// Create session
	svc, err := CloudFrontService(ctx, d)
	if err != nil {
		return nil, err
	}

	// List call
	result, err := svc.ListCachePolicies(&cloudfront.ListCachePoliciesInput{})

	for _, policy := range result.CachePolicyList.Items {
		d.StreamListItem(ctx, policy)
	}

	return nil, err
}

//// HYDRATE FUNCTIONS

func getCloudFrontCachePolicy(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getCloudFrontCachePolicy")

	// Create session
	svc, err := CloudFrontService(ctx, d)
	if err != nil {
		return nil, err
	}

	var id string
	if h.Item != nil {
		id = *h.Item.(*cloudfront.CachePolicySummary).CachePolicy.Id
	} else {
		id = d.KeyColumnQuals["id"].GetStringValue()
	}

	params := &cloudfront.GetCachePolicyInput{
		Id: aws.String(id),
	}

	op, err := svc.GetCachePolicy(params)
	if err != nil {
		return nil, err
	}

	return op, nil
}

func getCloudfrontCachePolicyAkas(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getCloudfrontCachePolicyAkas")
	id := cloudFrontCachePolicyAka(h.Item)
	getCommonColumnsCached := plugin.HydrateFunc(getCommonColumns).WithCache()
	commonData, err := getCommonColumnsCached(ctx, d, h)
	if err != nil {
		return nil, err
	}
	commonColumnData := commonData.(*awsCommonColumnData)

	akas := []string{"arn:" + commonColumnData.Partition + ":cloudfront::" + commonColumnData.AccountId + ":cache-policy/" + *id}

	return akas, nil
}

//// TRANSFORM FUNCTIONS

func cloudFrontCachePolicyAka(item interface{}) *string {
	switch item := item.(type) {
	case *cloudfront.GetCachePolicyOutput:
		return item.CachePolicy.Id
	case *cloudfront.CachePolicySummary:
		return item.CachePolicy.Id
	}
	return nil
}
