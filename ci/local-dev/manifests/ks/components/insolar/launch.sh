#@IgnoreInspection BashAddShebang
set -e
set -x
CONFIG_DIR=/opt/insolar/config
BOOTSTRAP_CONFIG=$CONFIG_DIR/bootstrap.yaml
HEAVY_GENESIS_CONFIG=$CONFIG_DIR/heavy_genesis.json
DISCOVERY_KEYS=$CONFIG_DIR/discovery
CERTS_KEYS=$CONFIG_DIR/certs

ls -alhR /opt
if [[ "$HOSTNAME" = "seed-0" && ! $(test -e /opt/insolar/config/finished) ]]
then
    echo "generate root member keys"
    insolar gen-key-pair > ${CONFIG_DIR}/root_member_keys.json

    echo "generate admin member_keys: $MD_ADMIN_MEMBER_KEYS_FILE"
    insolar gen-key-pair > ${CONFIG_DIR}/md_admin_member_keys.json

    echo "generate oracles member_keys"
    insolar gen-key-pair > ${CONFIG_DIR}/oracle0_member_keys.json
    insolar gen-key-pair > ${CONFIG_DIR}/oracle1_member_keys.json
    insolar gen-key-pair > ${CONFIG_DIR}/oracle2_member_keys.json
    insolar gen-key-pair > ${CONFIG_DIR}/oracle3_member_keys.json
    insolar gen-key-pair > ${CONFIG_DIR}/oracle4_member_keys.json
    insolar gen-key-pair > ${CONFIG_DIR}/oracle5_member_keys.json
    insolar gen-key-pair > ${CONFIG_DIR}/oracle6_member_keys.json
    insolar gen-key-pair > ${CONFIG_DIR}/oracle7_member_keys.json
    insolar gen-key-pair > ${CONFIG_DIR}/oracle8_member_keys.json
    insolar gen-key-pair > ${CONFIG_DIR}/oracle9_member_keys.json

    echo "generate bootstrap files"
    mkdir -vp $CERTS_KEYS
    mkdir -vp $DISCOVERY_KEYS
    insolar bootstrap --config ${BOOTSTRAP_CONFIG} --certificates-out-dir ${CERTS_KEYS}
    touch /opt/insolar/config/finished
else
    while ! (/usr/bin/test -e /opt/insolar/config/finished)
    do
        echo "Waiting for bootstrap ... ( sleep 5 sec )"
        sleep 5s
    done
fi

if [[ -f /opt/work/config/node-cert.json ]];
then
    echo "skip work"
else
    mkdir -vp /opt/work/config

    echo "copy files required for genesis"
    cp -v ${HEAVY_GENESIS_CONFIG} /opt/work/config/heavy_genesis.json
    cp -vR ${CONFIG_DIR}/plugins /opt/work/

    echo "copy root member keys for benchmarking purposes"
    cp -v ${CONFIG_DIR}/root_member_keys.json /opt/work/config/

    echo "copy configs"
    cp -v ${CERTS_KEYS}/$(hostname | awk -F'-' '{ printf "seed-%d-cert.json", $2 }')  /opt/work/config/node-cert.json
    cp -v ${DISCOVERY_KEYS}/$(hostname | awk -F'-' '{ printf "seed-%d-key.json", $2 }')  /opt/work/config/node-keys.json
fi
