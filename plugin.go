package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"code.cloudfoundry.org/cli/plugin"
)

type MultiCmd struct{}

func (c *MultiCmd) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "project42",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 1,
		},
		Commands: []plugin.Command{
			{
				Name:     "careless-delivery",
				Alias:    "cd",
				HelpText: "create a careless continuous delivery experience for an already pushed app",
				UsageDetails: plugin.Usage{
					Usage: "careless-delivery APP_NAME \n\nENVIRONMENT:\n    CF_USERNAME        Username for the pipeline to access the app deployed to Cloud Foundry\n    CF_PASSWORD        Password for the pipeline to access the app deployed to Cloud Foundry\n    CONCOURSE_TARGET     The concourse server to target\n    CONCOURSE_USERNAME Username to login to concourse server with the cf cli\n    CONCOURSE_PASSWORD Password to login to concourse server with the cf cli\n\nPREREQUESITES:\n    Installed fly and git CLIs on your machine.\n\nUSAGE EXAMPLE:\n    cf careless-delivery myapp\n\nSEE ALSO:\n    push",
				},
			},
			{
				Name:     "spin-up-prod",
				Alias:    "sup",
				HelpText: "create a careless continuous delivery experience to bring your application to production for an already pushed app",
				UsageDetails: plugin.Usage{
					Usage: "spin-up-prod APP_NAME \n\nENVIRONMENT:\n    CF_USERNAME        Username for the pipeline to access the app deployed to Cloud Foundry\n    CF_PASSWORD        Password for the pipeline to access the app deployed to Cloud Foundry\n    CONCOURSE_TARGET     The concourse server to target\n    CONCOURSE_USERNAME Username to login to concourse server with the cf cli\n    CONCOURSE_PASSWORD Password to login to concourse server with the cf cli\n\nPREREQUESITES:\n    Installed fly and git CLIs on your machine.\n\nUSAGE EXAMPLE:\n    cf spin-up-prod myapp\n\nSEE ALSO:\n    push",
					Options: map[string]string{
						"r": "Route of the app to use in production.",
						"o": "Organisation to deploy the productive app to (optional).",
						"s": "Space to deploy the productive app to (optional).",
					},
				},
			},
		},
	}
}

func main() {
	plugin.Start(new(MultiCmd))
}

func (c *MultiCmd) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "careless-delivery" {
		c.CarelessDelivery(args[1:])
	} else if args[0] == "spin-up-prod" {
		c.SpinUpProd(args[1:])
	}
}

func (c *MultiCmd) CarelessDelivery(args []string) {
	fmt.Println("Lean back and let me setup all for you...")
	script := "./careless-delivery.sh"
	fmt.Println(args[1:])
	c.ExecuteScript(script, carelessDeliverySh, args)
}

func (c *MultiCmd) SpinUpProd(args []string) {
	fmt.Println("Lean back and let me spin up all for you...")
	script := "./spin-up-prod.sh"
	fmt.Println(args[1:])
	c.ExecuteScript(script, spinUpProdSh, args)
}

func (c *MultiCmd) ExecuteScript(scriptName string, script string, args []string) {
	c.CreateScript(scriptName, script)
	cmd := exec.Command(scriptName, args...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		m := scanner.Text()
		log.Printf(m)
	}

	errScanner := bufio.NewScanner(stderr)
	for errScanner.Scan() {
		m := errScanner.Text()
		log.Printf(m)
	}

	cmd.Wait()

	// if _, err := os.Stat(scriptName); !os.IsNotExist(err) {
	// 	err := os.Remove(scriptName)

	_, err := os.Stat(scriptName)
	if err != nil {
		fmt.Println(err)
		return
	}
	// }
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func (c *MultiCmd) CreateScript(script string, code string) {
	data := []byte(code)
	err := ioutil.WriteFile(script, data, 0777)
	check(err)
}

var carelessDeliverySh = `#!/bin/bash
set -e -o pipefail

DEFAULT_BUILDPACK_PIPELINE_REPOSITORY="https://raw.githubusercontent.com/evoila/project42/develop"

function deploy() {
    local pipeline=$1
    local git_rev=$2
    local route=$3

    if [[ ! -d .cd ]]; then
        mkdir .cd
    fi

    BUILDPACK=$(cf curl /v2/buildpacks/$(cf curl "/v2/apps?q=name:${APP_NAME}" | jq '.resources[0].entity.detected_buildpack_guid' -r) | jq '.entity.name' -r)

    BUILDPACK_PIPELINE="${BUILDPACK}-pipeline.yml"

    PIPELINE=".cd/${BUILDPACK_PIPELINE}"
    #touch "${PIPELINE}"
    curl --silent "${BUILDPACK_PIPELINE_REPOSITORY:-$DEFAULT_BUILDPACK_PIPELINE_REPOSITORY}/${BUILDPACK_PIPELINE}" -o "${PIPELINE}"

    fly -t "${CONCOURSE_TARGET}" login -k -u "${CONCOURSE_USERNAME}" -p "${CONCOURSE_PASSWORD}"
    fly -t "${CONCOURSE_TARGET}" set-pipeline -n -p "${pipeline}" -c "${PIPELINE}" \
        -v cf_username="${CF_USERNAME}" \
        -v cf_password="${CF_PASSWORD}" \
        -v cf_api="${CF_API}" \
        -v app_name="${APP_NAME}" \
        -v git_url="${GIT_URL}" \
        -v cf_source_org="${CF_SOURCE_ORG}" \
        -v cf_source_space="${CF_SOURCE_SPACE}" \
        -v cf_target_org="${CF_TARGET_ORG}" \
        -v cf_target_space="${CF_TARGET_SPACE}" \
        -v git_rev="${git_rev}" \
		-v route="${route}"
}

GIT_URL=$(git remote -v | awk '{print $2}' | head -n 1)
APP_NAME="$1"
CF_API=$(cf t | grep "^api endpoint:" | awk '{print $3}')
CF_SOURCE_ORG=$(cf t | grep "^org:" | awk '{print $2}')
CF_SOURCE_SPACE=$(cf t | grep "^space:" | awk '{print $2}')
CF_TARGET_ORG="${CF_SOURCE_ORG}"
CF_TARGET_SPACE="${CF_SOURCE_SPACE}"

GIT_COMMIT=$(git rev-parse HEAD)

deploy "${APP_NAME}" "${GIT_COMMIT}" ""
`
var spinUpProdSh = `#!/bin/bash
set -e -o pipefail
APP_NAME="$1"

DEFAULT_BUILDPACK_PIPELINE_REPOSITORY="https://raw.githubusercontent.com/evoila/project42/develop"

function deploy() {
    local pipeline=$1
    local git_rev=$2
    local route=$3

    if [[ ! -d .cd ]]; then
        mkdir .cd
    fi

    BUILDPACK=$(cf curl /v2/buildpacks/$(cf curl "/v2/apps?q=name:${APP_NAME}" | jq '.resources[0].entity.detected_buildpack_guid' -r) | jq '.entity.name' -r)

    BUILDPACK_PIPELINE="${BUILDPACK}-pipeline.yml"

    PIPELINE=".cd/${BUILDPACK_PIPELINE}"
    #touch "${PIPELINE}"
    curl --silent "${BUILDPACK_PIPELINE_REPOSITORY:-$DEFAULT_BUILDPACK_PIPELINE_REPOSITORY}/${BUILDPACK_PIPELINE}" -o "${PIPELINE}"

    fly -t "${CONCOURSE_TARGET}" login -k -u "${CONCOURSE_USERNAME}" -p "${CONCOURSE_PASSWORD}"
    fly -t "${CONCOURSE_TARGET}" set-pipeline -n -p "${pipeline}" -c "${PIPELINE}" \
        -v cf_username="${CF_USERNAME}" \
        -v cf_password="${CF_PASSWORD}" \
        -v cf_api="${CF_API}" \
        -v app_name="${APP_NAME}" \
        -v git_url="${GIT_URL}" \
        -v cf_source_org="${CF_SOURCE_ORG}" \
        -v cf_source_space="${CF_SOURCE_SPACE}" \
        -v cf_target_org="${CF_TARGET_ORG}" \
        -v cf_target_space="${CF_TARGET_SPACE}" \
        -v git_rev="${git_rev}" \
		-v route="${route}"
}

if [[ $2 == '-o' ]]; then
	CF_TARGET_ORG=$3
	shift 2
else
	CF_TARGET_ORG=$(cf t | grep "^org:" | awk '{print $2}')
fi

if [[ $2 == '-s' ]]; then
	CF_TARGET_SPACE=$3
	shift 2
else
	CF_TARGET_SPACE=$(cf t | grep "^space:" | awk '{print $2}')
	APP_NAME="${1}-prod"
fi

if [[ $2 == '-r' ]]; then
	ROUTE=$3
	shift 2
else
	echo "$(tput setaf 1)$(tput bold)Route has to be provided using -r$(tput sgr0)" >&2
	exit 2
fi

GIT_URL=$(git remote -v | awk '{print $2}' | head -n 1)
CF_API=$(cf t | grep "^api endpoint:" | awk '{print $3}')
CF_SOURCE_ORG=$(cf t | grep "^org:" | awk '{print $2}')
CF_SOURCE_SPACE=$(cf t | grep "^space:" | awk '{print $2}')

GIT_REV=$(cf curl "/v2/apps?q=name:${APP_NAME}" | jq '.resources[0].entity.environment_json.GIT_REV' -r)
if [[ -z "${GIT_REV}" || "${GIT_REV}" == "null" ]]; then
	GIT_REV=$(git rev-parse HEAD)
fi

deploy "${APP_NAME}-prod" "${GIT_REV}" "${ROUTE}"
`
