#!/usr/bin/env groovy

pipeline {
  agent { label 'conjur-enterprise-common-agent' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
  }

  triggers {
    cron(getDailyCronString())
  }

  stages {
    stage('Get InfraPool ExecutorV2 Agent') {
      steps {
        script {
          // Get InfraPool ExecutorV2 Agent
          INFRAPOOL_EXECUTORV2_AGENT_0 = getInfraPoolAgent.connected(type: "ExecutorV2", quantity: 1, duration: 1)[0]
        }
      }
    }

    stage('Release test') {
      steps {
        script {
          INFRAPOOL_EXECUTORV2_AGENT_0.agentSh 'summon -f ./k8s-ci/secrets.yml ./k8s-ci/test release-testing'
        }
      }
    }
  }

  post {
    always {
      releaseInfraPoolAgent(".infrapool/release_agents")
     }
  }
}
