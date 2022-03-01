# Customize your jitsi web interface

## How to build your own manager

### Prerequisites : 
- install go 1.15.5
- install controller-gen
- install gcc and build-base

### Steps :

    * go to Makefile: change IMG ?= with your own repository docker 
    * make build
    * make docker-build
    * make docker-push 
    * make generate-deploy
    * make deploy
    * kubectl apply -f deploy/jitsi-operator
- Install your configmap and your Jitsi stack (cf config/samples/mydomain.com_jitsi_instance.yaml  config/samples/mydomain.com_jitsi_instance.yaml)

## Custom interface

- Explanations of some parameters

In the config.js file, add config.dynamicBrandingUrl, this parameter is usefull  to change many style configuration in jitsi .

```bash
config.dynamicBrandingUrl = "https://meet/jitsi/dynamicBranding.json";
```
cf https://community.jitsi.org/t/queries-about-dynamic-branding-url/101702

Set up these parameter in json file which must be publically available
An example of json file below:

```bash

{
	"backgroundImageUrl": "YOUR_LINK_BACKGROUNDIMAGE",
	"premeetingBackground": "url(YOUR_LINK_PREMEETING_BACKGROUNDIMAGE)",
	"virtualBackgrounds": [ "VIRTUAL_BACKGROUNDIMAGE_1", 
							"VIRTUAL_BACKGROUNDIMAGE_2",
							"VIRTUAL_BACKGROUNDIMAGE_3",
							"VIRTUAL_BACKGROUNDIMAGE_4"
	]
}


```
* "backgroundImageUrl": link of background image ( during conference )
* "premeetingBackground": link of background image for premeeting page (where you enter your name to join the meeting)
* "virtualBackgrounds": link of virtual background images 


