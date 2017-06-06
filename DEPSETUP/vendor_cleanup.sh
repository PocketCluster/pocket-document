#!/usr/bin/env bash

function clean_vendor() {
	rm -rf github.com/docker/go-units && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf github.com/hashicorp/serf && (rmdir github.com/hashicorp > /dev/null 2>&1 || true)
	rm -rf golang.org/x/crypto && (rmdir golang.org/x > /dev/null 2>&1 || true)
	rm -rf github.com/docker/leadership && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf github.com/stretchr/objx && (rmdir github.com/stretchr > /dev/null 2>&1 || true)
	rm -rf github.com/gogo/protobuf && (rmdir github.com/gogo > /dev/null 2>&1 || true)
	rm -rf gopkg.in/check.v1 && (rmdir gopkg.in > /dev/null 2>&1 || true)
	rm -rf golang.org/x/sys && (rmdir golang.org/x > /dev/null 2>&1 || true)
	rm -rf github.com/docker/libkv && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf github.com/bshuster-repo/logrus-logstash-hook && (rmdir github.com/bshuster-repo > /dev/null 2>&1 || true)
	rm -rf github.com/stretchr/testify && (rmdir github.com/stretchr > /dev/null 2>&1 || true)
	rm -rf github.com/samuel/go-zookeeper && (rmdir github.com/samuel > /dev/null 2>&1 || true)
	rm -rf github.com/coreos/pkg && (rmdir github.com/coreos > /dev/null 2>&1 || true)
	rm -rf github.com/prometheus/client_model && (rmdir github.com/prometheus > /dev/null 2>&1 || true)
	rm -rf github.com/cpuguy83/go-md2man && (rmdir github.com/cpuguy83 > /dev/null 2>&1 || true)
	rm -rf github.com/inconshreveable/mousetrap && (rmdir github.com/inconshreveable > /dev/null 2>&1 || true)
	rm -rf google.golang.org/api && (rmdir google.golang.org > /dev/null 2>&1 || true)
	rm -rf github.com/pborman/uuid && (rmdir github.com/pborman > /dev/null 2>&1 || true)
	rm -rf github.com/coreos/go-systemd && (rmdir github.com/coreos > /dev/null 2>&1 || true)
	rm -rf github.com/matttproud/golang_protobuf_extensions && (rmdir github.com/matttproud > /dev/null 2>&1 || true)
	rm -rf github.com/cloudflare/cfssl && (rmdir github.com/cloudflare > /dev/null 2>&1 || true)
	rm -rf github.com/docker/docker && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf google.golang.org/grpc && (rmdir google.golang.org > /dev/null 2>&1 || true)
	rm -rf golang.org/x/oauth2 && (rmdir golang.org/x > /dev/null 2>&1 || true)
	rm -rf github.com/mattn/go-sqlite3 && (rmdir github.com/mattn > /dev/null 2>&1 || true)
	rm -rf github.com/docker/libtrust && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf github.com/ugorji/go && (rmdir github.com/ugorji > /dev/null 2>&1 || true)
	rm -rf github.com/pkg/errors && (rmdir github.com/pkg > /dev/null 2>&1 || true)
	rm -rf golang.org/x/time && (rmdir golang.org/x > /dev/null 2>&1 || true)
	rm -rf github.com/coreos/etcd && (rmdir github.com/coreos > /dev/null 2>&1 || true)
	rm -rf github.com/spf13/cobra && (rmdir github.com/spf13 > /dev/null 2>&1 || true)
	rm -rf github.com/Azure/go-ansiterm && (rmdir github.com/Azure > /dev/null 2>&1 || true)
	rm -rf golang.org/x/net && (rmdir golang.org/x > /dev/null 2>&1 || true)
	rm -rf github.com/gorilla/context && (rmdir github.com/gorilla > /dev/null 2>&1 || true)
	rm -rf github.com/mitchellh/mapstructure && (rmdir github.com/mitchellh > /dev/null 2>&1 || true)
	rm -rf github.com/docker/distribution && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf github.com/pmezard/go-difflib && (rmdir github.com/pmezard > /dev/null 2>&1 || true)
	rm -rf github.com/prometheus/common && (rmdir github.com/prometheus > /dev/null 2>&1 || true)
	rm -rf github.com/golang/protobuf && (rmdir github.com/golang > /dev/null 2>&1 || true)
	rm -rf github.com/jonboulle/clockwork && (rmdir github.com/jonboulle > /dev/null 2>&1 || true)
	rm -rf github.com/prometheus/client_golang && (rmdir github.com/prometheus > /dev/null 2>&1 || true)
	rm -rf github.com/mattn/go-shellwords && (rmdir github.com/mattn > /dev/null 2>&1 || true)
	rm -rf github.com/spf13/pflag && (rmdir github.com/spf13 > /dev/null 2>&1 || true)
	rm -rf github.com/prometheus/procfs && (rmdir github.com/prometheus > /dev/null 2>&1 || true)
	rm -rf github.com/urfave/cli && (rmdir github.com/urfave > /dev/null 2>&1 || true)
	rm -rf gopkg.in/yaml.v2 && (rmdir gopkg.in > /dev/null 2>&1 || true)
	rm -rf github.com/opencontainers/runc && (rmdir github.com/opencontainers > /dev/null 2>&1 || true)
	rm -rf github.com/vbatts/tar-split && (rmdir github.com/vbatts > /dev/null 2>&1 || true)
	rm -rf github.com/gorilla/mux && (rmdir github.com/gorilla > /dev/null 2>&1 || true)
	rm -rf github.com/beorn7/perks && (rmdir github.com/beorn7 > /dev/null 2>&1 || true)
	rm -rf github.com/Sirupsen/logrus && (rmdir github.com/Sirupsen > /dev/null 2>&1 || true)
	rm -rf github.com/docker/go-connections && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf github.com/xeipuuv/gojsonschema && (rmdir github.com/xeipuuv > /dev/null 2>&1 || true)
	rm -rf github.com/xeipuuv/gojsonreference && (rmdir github.com/xeipuuv > /dev/null 2>&1 || true)
	rm -rf github.com/godbus/dbus && (rmdir github.com/godbus > /dev/null 2>&1 || true)
	rm -rf github.com/miekg/dns && (rmdir github.com/miekg > /dev/null 2>&1 || true)
	rm -rf github.com/davecgh/go-spew && (rmdir github.com/davecgh > /dev/null 2>&1 || true)
	rm -rf github.com/Microsoft/go-winio && (rmdir github.com/Microsoft > /dev/null 2>&1 || true)
	rm -rf github.com/boltdb/bolt && (rmdir github.com/boltdb > /dev/null 2>&1 || true)
	rm -rf github.com/kr/pty && (rmdir github.com/kr > /dev/null 2>&1 || true)
	rm -rf github.com/hashicorp/consul && (rmdir github.com/hashicorp > /dev/null 2>&1 || true)
	rm -rf github.com/xeipuuv/gojsonpointer && (rmdir github.com/xeipuuv > /dev/null 2>&1 || true)
	rm -rf github.com/coreos/etcd && (rmdir github.com/coreos > /dev/null 2>&1 || true)
	rm -rf github.com/docker/libcompose && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf github.com/docker/swarm && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf github.com/docker/distribution && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf github.com/docker/docker && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf github.com/gravitational/teleport && (rmdir github.com/gravitational > /dev/null 2>&1 || true)
}

function clean_gopath() {
	rm -rf github.com/mistifyio/go-zfs && (rmdir github.com/mistifyio > /dev/null 2>&1 || true)
	rm -rf github.com/mailgun/minheap && (rmdir github.com/mailgun > /dev/null 2>&1 || true)
	rm -rf github.com/bsphere/le_go && (rmdir github.com/bsphere > /dev/null 2>&1 || true)
	rm -rf github.com/agl/ed25519 && (rmdir github.com/agl > /dev/null 2>&1 || true)
	rm -rf github.com/yvasiyarov/go-metrics && (rmdir github.com/yvasiyarov > /dev/null 2>&1 || true)
	rm -rf github.com/golang/mock && (rmdir github.com/golang > /dev/null 2>&1 || true)
	rm -rf github.com/hashicorp/go-cleanhttp && (rmdir github.com/hashicorp > /dev/null 2>&1 || true)
	rm -rf github.com/mailgun/ttlmap && (rmdir github.com/mailgun > /dev/null 2>&1 || true)
	rm -rf github.com/docker/containerd && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf github.com/aws/aws-sdk-go && (rmdir github.com/aws > /dev/null 2>&1 || true)
	rm -rf github.com/olekukonko/tablewriter && (rmdir github.com/olekukonko > /dev/null 2>&1 || true)
	rm -rf github.com/docker/swarmkit && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf github.com/vulcand/oxy && (rmdir github.com/vulcand > /dev/null 2>&1 || true)
	rm -rf github.com/google/certificate-transparency && (rmdir github.com/google > /dev/null 2>&1 || true)
	rm -rf github.com/kardianos/osext && (rmdir github.com/kardianos > /dev/null 2>&1 || true)
	rm -rf github.com/miekg/pkcs11 && (rmdir github.com/miekg > /dev/null 2>&1 || true)
	rm -rf github.com/mailgun/lemma && (rmdir github.com/mailgun > /dev/null 2>&1 || true)
	rm -rf github.com/garyburd/redigo && (rmdir github.com/garyburd > /dev/null 2>&1 || true)
	rm -rf github.com/grpc-ecosystem/grpc-gateway && (rmdir github.com/grpc-ecosystem > /dev/null 2>&1 || true)
	rm -rf github.com/julienschmidt/httprouter && (rmdir github.com/julienschmidt > /dev/null 2>&1 || true)
	rm -rf github.com/bugsnag/panicwrap && (rmdir github.com/bugsnag > /dev/null 2>&1 || true)
	rm -rf github.com/mailgun/timetools && (rmdir github.com/mailgun > /dev/null 2>&1 || true)
	rm -rf github.com/bugsnag/bugsnag-go && (rmdir github.com/bugsnag > /dev/null 2>&1 || true)
	rm -rf github.com/docker/docker-credential-helpers && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf github.com/karlseguin/ccache && (rmdir github.com/karlseguin > /dev/null 2>&1 || true)
	rm -rf github.com/mdp/rsc && (rmdir github.com/mdp > /dev/null 2>&1 || true)
	rm -rf github.com/tstranex/u2f && (rmdir github.com/tstranex > /dev/null 2>&1 || true)
	rm -rf github.com/opencontainers/runtime-spec && (rmdir github.com/opencontainers > /dev/null 2>&1 || true)
	rm -rf github.com/BurntSushi/toml && (rmdir github.com/BurntSushi > /dev/null 2>&1 || true)
	rm -rf github.com/Graylog2/go-gelf && (rmdir github.com/Graylog2 > /dev/null 2>&1 || true)
	rm -rf github.com/hashicorp/go-immutable-radix && (rmdir github.com/hashicorp > /dev/null 2>&1 || true)
	rm -rf github.com/vishvananda/netns && (rmdir github.com/vishvananda > /dev/null 2>&1 || true)
	rm -rf github.com/go-ini/ini && (rmdir github.com/go-ini > /dev/null 2>&1 || true)
	rm -rf github.com/coreos/go-semver && (rmdir github.com/coreos > /dev/null 2>&1 || true)
	rm -rf github.com/coreos/go-oidc && (rmdir github.com/coreos > /dev/null 2>&1 || true)
	rm -rf bitbucket.org/ww/goautoneg && (rmdir bitbucket.org/ww > /dev/null 2>&1 || true)
	rm -rf github.com/flynn/go-shlex && (rmdir github.com/flynn > /dev/null 2>&1 || true)
	rm -rf github.com/mitchellh/go-ps && (rmdir github.com/mitchellh > /dev/null 2>&1 || true)
	rm -rf github.com/google/btree && (rmdir github.com/google > /dev/null 2>&1 || true)
	rm -rf rsc.io/letsencrypt && (rmdir rsc.io > /dev/null 2>&1 || true)
	rm -rf github.com/alecthomas/template && (rmdir github.com/alecthomas > /dev/null 2>&1 || true)
	rm -rf github.com/fsnotify/fsnotify && (rmdir github.com/fsnotify > /dev/null 2>&1 || true)
	rm -rf github.com/jinzhu/gorm && (rmdir github.com/jinzhu > /dev/null 2>&1 || true)
	rm -rf github.com/docker/notary && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf k8s.io/client-go/1.4 && (rmdir k8s.io/client-go > /dev/null 2>&1 || true)
	rm -rf github.com/stevvooe/resumable && (rmdir github.com/stevvooe > /dev/null 2>&1 || true)
	rm -rf github.com/xiang90/probing && (rmdir github.com/xiang90 > /dev/null 2>&1 || true)
	rm -rf github.com/bugsnag/osext && (rmdir github.com/bugsnag > /dev/null 2>&1 || true)
	rm -rf github.com/go-check/check && (rmdir github.com/go-check > /dev/null 2>&1 || true)
	rm -rf github.com/hashicorp/go-memdb && (rmdir github.com/hashicorp > /dev/null 2>&1 || true)
	rm -rf github.com/gravitational/kingpin && (rmdir github.com/gravitational > /dev/null 2>&1 || true)
	rm -rf github.com/imdario/mergo && (rmdir github.com/imdario > /dev/null 2>&1 || true)
	rm -rf github.com/mailgun/oxy && (rmdir github.com/mailgun > /dev/null 2>&1 || true)
	rm -rf github.com/golang/glog && (rmdir github.com/golang > /dev/null 2>&1 || true)
	rm -rf github.com/yvasiyarov/gorelic && (rmdir github.com/yvasiyarov > /dev/null 2>&1 || true)
	rm -rf github.com/hashicorp/go-msgpack && (rmdir github.com/hashicorp > /dev/null 2>&1 || true)
	rm -rf github.com/flynn-archive/go-shlex && (rmdir github.com/flynn-archive > /dev/null 2>&1 || true)
	rm -rf github.com/mreiferson/go-httpclient && (rmdir github.com/mreiferson > /dev/null 2>&1 || true)
	rm -rf golang.org/x/text && (rmdir golang.org/x > /dev/null 2>&1 || true)
	rm -rf github.com/codegangsta/cli && (rmdir github.com/codegangsta > /dev/null 2>&1 || true)
	rm -rf github.com/hashicorp/go-multierror && (rmdir github.com/hashicorp > /dev/null 2>&1 || true)
	rm -rf github.com/ricochet2200/go-disk-usage && (rmdir github.com/ricochet2200 > /dev/null 2>&1 || true)
	rm -rf github.com/yvasiyarov/newrelic_platform_go && (rmdir github.com/yvasiyarov > /dev/null 2>&1 || true)
	rm -rf github.com/dustin/go-humanize && (rmdir github.com/dustin > /dev/null 2>&1 || true)
	rm -rf github.com/deckarep/golang-set && (rmdir github.com/deckarep > /dev/null 2>&1 || true)
	rm -rf github.com/tinylib/msgp && (rmdir github.com/tinylib > /dev/null 2>&1 || true)
	rm -rf github.com/pivotal-golang/clock && (rmdir github.com/pivotal-golang > /dev/null 2>&1 || true)
	rm -rf github.com/fluent/fluent-logger-golang && (rmdir github.com/fluent > /dev/null 2>&1 || true)
	rm -rf google.golang.org/appengine && (rmdir google.golang.org > /dev/null 2>&1 || true)
	rm -rf github.com/jmoiron/sqlx && (rmdir github.com/jmoiron > /dev/null 2>&1 || true)
	rm -rf github.com/gravitational/ttlmap && (rmdir github.com/gravitational > /dev/null 2>&1 || true)
	rm -rf github.com/gravitational/trace && (rmdir github.com/gravitational > /dev/null 2>&1 || true)
	rm -rf github.com/gorilla/handlers && (rmdir github.com/gorilla > /dev/null 2>&1 || true)
	rm -rf github.com/syndtr/gocapability && (rmdir github.com/syndtr > /dev/null 2>&1 || true)
	rm -rf github.com/bgentry/speakeasy && (rmdir github.com/bgentry > /dev/null 2>&1 || true)
	rm -rf github.com/tchap/go-patricia && (rmdir github.com/tchap > /dev/null 2>&1 || true)
	rm -rf github.com/hashicorp/golang-lru && (rmdir github.com/hashicorp > /dev/null 2>&1 || true)
	rm -rf github.com/docker/go-events && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf github.com/mesos/mesos-go && (rmdir github.com/mesos > /dev/null 2>&1 || true)
	rm -rf gopkg.in/cheggaaa/pb.v1 && (rmdir gopkg.in/cheggaaa > /dev/null 2>&1 || true)
	rm -rf github.com/samalba/dockerclient && (rmdir github.com/samalba > /dev/null 2>&1 || true)
	rm -rf github.com/skarademir/naturalsort && (rmdir github.com/skarademir > /dev/null 2>&1 || true)
	rm -rf github.com/seccomp/libseccomp-golang && (rmdir github.com/seccomp > /dev/null 2>&1 || true)
	rm -rf github.com/philhofer/fwd && (rmdir github.com/philhofer > /dev/null 2>&1 || true)
	rm -rf github.com/jmespath/go-jmespath && (rmdir github.com/jmespath > /dev/null 2>&1 || true)
	rm -rf github.com/docker/go && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf github.com/alecthomas/units && (rmdir github.com/alecthomas > /dev/null 2>&1 || true)
	rm -rf github.com/gravitational/configure && (rmdir github.com/gravitational > /dev/null 2>&1 || true)
	rm -rf github.com/ghodss/yaml && (rmdir github.com/ghodss > /dev/null 2>&1 || true)
	rm -rf github.com/mailgun/metrics && (rmdir github.com/mailgun > /dev/null 2>&1 || true)
	rm -rf github.com/vdemeester/shakers && (rmdir github.com/vdemeester > /dev/null 2>&1 || true)
	rm -rf github.com/docker/go-metrics && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf github.com/vishvananda/netlink && (rmdir github.com/vishvananda > /dev/null 2>&1 || true)
	rm -rf github.com/armon/go-metrics && (rmdir github.com/armon > /dev/null 2>&1 || true)
	rm -rf github.com/cockroachdb/cmux && (rmdir github.com/cockroachdb > /dev/null 2>&1 || true)
	rm -rf github.com/jinzhu/inflection && (rmdir github.com/jinzhu > /dev/null 2>&1 || true)
	rm -rf github.com/Microsoft/hcsshim && (rmdir github.com/Microsoft > /dev/null 2>&1 || true)
	rm -rf github.com/grpc-ecosystem/go-grpc-prometheus && (rmdir github.com/grpc-ecosystem > /dev/null 2>&1 || true)
	rm -rf github.com/docker/libnetwork && (rmdir github.com/docker > /dev/null 2>&1 || true)
	rm -rf github.com/gravitational/roundtrip && (rmdir github.com/gravitational > /dev/null 2>&1 || true)
	rm -rf github.com/aanand/compose-file && (rmdir github.com/aanand > /dev/null 2>&1 || true)
	rm -rf github.com/tonistiigi/fifo && (rmdir github.com/tonistiigi > /dev/null 2>&1 || true)
	rm -rf github.com/gokyle/hotp && (rmdir github.com/gokyle > /dev/null 2>&1 || true)
	rm -rf gopkg.in/vmihailenco/msgpack.v2 && (rmdir gopkg.in/vmihailenco > /dev/null 2>&1 || true)
	rm -rf github.com/hashicorp/memberlist && (rmdir github.com/hashicorp > /dev/null 2>&1 || true)
	rm -rf github.com/spf13/afero && (rmdir github.com/spf13 > /dev/null 2>&1 || true)
	rm -rf google.golang.org/cloud && (rmdir google.golang.org > /dev/null 2>&1 || true)
	rm -rf github.com/armon/go-radix && (rmdir github.com/armon > /dev/null 2>&1 || true)
	rm -rf github.com/getsentry/raven-go && (rmdir github.com/getsentry > /dev/null 2>&1 || true)
	rm -rf github.com/RackSec/srslog && (rmdir github.com/RackSec > /dev/null 2>&1 || true)
	rm -rf github.com/buger/goterm && (rmdir github.com/buger > /dev/null 2>&1 || true)
}
