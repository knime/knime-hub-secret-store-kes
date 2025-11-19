#!groovy

// Jenkinsfile for building an application artifact and a docker image.

// load knime library depending on the branchname.
def BN = BRANCH_NAME == "main" ? "master" : "releases/2023-12"
library "knime-pipeline@$BN"

properties([
    buildDiscarder(logRotator(numToKeepStr: '5')),
    disableConcurrentBuilds(),
    parameters([
            booleanParam(defaultValue: false, description: 'Whether this is a release build', name: 'RELEASE_BUILD'),
            booleanParam(defaultValue: false, description: 'Whether to force deployment of the docker image', name: 'FORCE_DEPLOYMENT')
        ])
])

timeout(time: 15, unit: 'MINUTES') {
    node('docker') {
        dockerTools.ecrLogin()

        try {
            knimetools.golangBuild()
        } catch (ex) {
            currentBuild.result = 'FAILURE'
            throw ex
        } finally {
            // no frequent changes expected, no build notifications required
        }
    }
}
