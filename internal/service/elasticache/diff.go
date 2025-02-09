// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/service/elasticache"
	gversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var minMemcachedTransitEncryptionVersion = gversion.Must(gversion.NewVersion("1.6.12"))

// CustomizeDiffValidateClusterAZMode validates that `num_cache_nodes` is greater than 1 when `az_mode` is "cross-az"
func CustomizeDiffValidateClusterAZMode(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if v, ok := diff.GetOk("az_mode"); !ok || v.(string) != elasticache.AZModeCrossAz {
		return nil
	}
	if v, ok := diff.GetOk("num_cache_nodes"); !ok || v.(int) != 1 {
		return nil
	}

	return errors.New(`az_mode "cross-az" is not supported with num_cache_nodes = 1`)
}

// CustomizeDiffValidateClusterNumCacheNodes validates that `num_cache_nodes` is 1 when `engine` is "redis"
func CustomizeDiffValidateClusterNumCacheNodes(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if v, ok := diff.GetOk("engine"); !ok || v.(string) == engineMemcached {
		return nil
	}

	if v, ok := diff.GetOk("num_cache_nodes"); !ok || v.(int) == 1 {
		return nil
	}
	return errors.New(`engine "redis" does not support num_cache_nodes > 1`)
}

// CustomizeDiffClusterMemcachedNodeType causes re-creation when `node_type` is changed and `engine` is "memcached"
func CustomizeDiffClusterMemcachedNodeType(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	// Engine memcached does not currently support vertical scaling
	// https://docs.aws.amazon.com/AmazonElastiCache/latest/mem-ug/Scaling.html#Scaling.Memcached.Vertically
	if diff.Id() == "" || !diff.HasChange("node_type") {
		return nil
	}
	if v, ok := diff.GetOk("engine"); !ok || v.(string) == engineRedis {
		return nil
	}
	return diff.ForceNew("node_type")
}

// CustomizeDiffValidateClusterMemcachedSnapshotIdentifier validates that `final_snapshot_identifier` is not set when `engine` is "memcached"
func CustomizeDiffValidateClusterMemcachedSnapshotIdentifier(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if v, ok := diff.GetOk("engine"); !ok || v.(string) == engineRedis {
		return nil
	}
	if _, ok := diff.GetOk("final_snapshot_identifier"); !ok {
		return nil
	}
	return errors.New(`engine "memcached" does not support final_snapshot_identifier`)
}

// CustomizeDiffValidateReplicationGroupAutomaticFailover validates that `automatic_failover_enabled` is set when `multi_az_enabled` is true
func CustomizeDiffValidateReplicationGroupAutomaticFailover(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if v := diff.Get("multi_az_enabled").(bool); !v {
		return nil
	}
	if v := diff.Get("automatic_failover_enabled").(bool); !v {
		return errors.New(`automatic_failover_enabled must be true if multi_az_enabled is true`)
	}
	return nil
}

// CustomizeDiffValidateTransitEncryptionEnabled validates that an appropriate engine type and version
// are utilized when in-transit encryption is enabled
func CustomizeDiffValidateTransitEncryptionEnabled(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	if v, ok := diff.GetOk("transit_encryption_enabled"); ok && v.(bool) {
		if engine := diff.Get("engine").(string); engine == engineRedis {
			return errors.New("aws_elasticache_cluster does not support transit encryption using the redis engine, use aws_elasticache_replication_group instead")
		}

		engineVersion, ok := diff.GetOk("engine_version")
		if !ok {
			return nil
		}
		version, err := normalizeEngineVersion(engineVersion.(string))
		if err != nil {
			return err
		}
		if version.LessThan(minMemcachedTransitEncryptionVersion) {
			return fmt.Errorf("Transit encryption is not supported for memcached version %v", version)
		}
	}

	return nil
}
