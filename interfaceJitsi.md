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

## Customize the web interface

In the config.js file, add config.dynamicBrandingUrl, this parameter is useful to change many style configurations.

```bash
config.dynamicBrandingUrl = "https://meet/jitsi/dynamicBranding.json";
```
cf https://community.jitsi.org/t/queries-about-dynamic-branding-url/101702

Set up these parameters in your dynamicBrandingUrl json file, which must be publically available.
An example of such a json file below :

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
* "backgroundImageUrl": background image ( during conference )
* "premeetingBackground": background image of premeeting page (where you enter your name to join the meeting)
* "virtualBackgrounds": list of suggested virtual background images 


