YoLocation
==========

YoLocation is a sample project for using location data with Yo.

High level overview:

  - Written in Go (golang.org)
  - Uses Revel web framework (revel.github.com)
  - Autobuild docker container (mtsgrd/yolocation)

Try it out!
-----------

Send a Yo to STARBUCKSMAP to get a yoback link to the nearest Starbucks that's currently open.

To run your own instance, consider using a container-optimized Google Compute Engine image: https://cloud.google.com/compute/docs/containers/container_vms.

Howto
-----

containers.yaml
```
$ cat > containers.yaml
version: v1beta2
containers:
- name: yolocation
  image: mtsgrd/yolocation
  ports:
  - name: http
    hostPort: [external port]
    containerPort: 8080
  env:
  - name: YO_API_URL
    value: http://www.justyo.co/yo/
  - name: YO_API_TOKEN
    value: string
  - name: GOOGLE_API_KEY
    value: string (must have Places API enabled)
^CTRL-D

$ gcloud compute instances create [new server name] \
    --project [project name]
    --image container-vm-v20140925
    --image-project google-containers
    --metadata-from-file google-container-manifest=containers.yaml
    --tags http-server
    --zone us-central1-a
    --machine-type f1-micro
```

**Note**: You must manually open the external port in the gcloud firewall.

Yo API Account
--------------

The Yo API account configuration determines the query used against the Google Places API. To configure the account to ping back the closest Starbucks location, use the following:
```
CALLBACK URL: http://[xxx.xxx.xxx.xxx]:9000/starbucks/
```

Version
----

0.1

License
----

MIT


**Free Software, Hell Yeah!**
