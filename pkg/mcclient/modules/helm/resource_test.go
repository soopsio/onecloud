package helm

import (
	"encoding/json"
	"testing"
)

const (
	resStr = `==> v1/ServiceAccount
NAME             SECRETS  AGE
mon-kafka        1        3d
mon-kairosdb     1        3d
mon-servicetree  1        3d

==> v1beta1/RoleBinding
NAME             AGE
mon-kafka        3d
mon-kairosdb     3d
mon-servicetree  3d

==> v1beta2/Deployment
NAME                DESIRED  CURRENT  UP-TO-DATE  AVAILABLE  AGE
mon-kairosdb        1        1        1           1          3d
mon-monitor-stream  1        1        1           1          3d
mon-servicetree     1        1        1           1          3d

==> v1/Pod(related)
NAME                                 READY  STATUS   RESTARTS  AGE
mon-kairosdb-7bf9f5655d-z4mkt        1/1    Running  0         3d
mon-monitor-stream-5859d6bbbb-cr4l4  1/1    Running  0         2d
mon-servicetree-7f59fd6bfb-f8bmk     1/1    Running  0         3d
mon-zookeeper-0                      1/1    Running  1         3d
mon-zookeeper-1                      1/1    Running  0         2d
mon-zookeeper-2                      1/1    Running  1         3d
mon-kafka-0                          1/1    Running  0         2d
mon-kafka-1                          1/1    Running  0         3d
mon-kafka-2                          1/1    Running  0         3d
mon-cassandra-0                      1/1    Running  0         3d
mon-cassandra-1                      1/1    Running  0         2d
mon-cassandra-2                      1/1    Running  0         3d

==> v1beta1/PodDisruptionBudget
NAME           MIN AVAILABLE  MAX UNAVAILABLE  ALLOWED DISRUPTIONS  AGE
mon-zookeeper  N/A            1                1                    3d`
)

func TestParseResources(t *testing.T) {
	ret := ParseResources(resStr)
	bs, _ := json.MarshalIndent(ret, "", "  ")
	t.Logf("ret: %s", string(bs))
}
