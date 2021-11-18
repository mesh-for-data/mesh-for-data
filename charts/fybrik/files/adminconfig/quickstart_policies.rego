package adminconfig

# configure where transformations take place
config[{"transform": decision}] {
    policy := {"ID": "transform-geo", "description":"Governance based transformations must take place in the geography where the data is stored"}
    clusters := [ data.clusters[i].name | data.clusters[i].metadata.region == input.request.dataset.geography ]
    decision := {"policy": policy, "restrictions": {"clusters": clusters}}
}

# configure the scope of the read capability
config[{"read": decision}] {
    input.request.usage.read == true
    policy := {"ID": "read-scope", "description":"Deploy read at the workload scope"}
    decision := {"policy": policy, "restrictions": {"modules": {"capabilities.scope" : "workload"}}}
}

# configure where the read capability will be deployed
config[{"read": decision}] {
    input.request.usage.read == true
    policy := {"ID": "read-location", "description":"Deploy read in the workload cluster"}
    decision := {"policy": policy, "restrictions": {"clusters": [ input.workload.cluster.name]}}
}

# copy requested by the user
config[{"copy": decision}] {
    input.request.usage.copy == true
    policy := {"ID": "copy-request", "description":"Copy capability is requested by the user"}
    decision := {"policy": policy, "deploy": true}
}

# configure when implicit copies should be made
config[{"copy": decision}] {
    input.request.usage.read == true
    input.request.dataset.geography != input.workload.cluster.metadata.region
    count(input.actions) > 0
    clusters :=  [ data.clusters[i].name | data.clusters[i].metadata.region == input.request.dataset.geography ]
    policy := {"ID": "copy-remote", "description":"Implicit copies should be used if the data is in a different region than the compute, and transformations are required"}
    decision := {"policy": policy, "deploy": true, "restrictions": {"clusters": clusters}}
}
