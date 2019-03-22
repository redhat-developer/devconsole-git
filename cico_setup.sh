#!/bin/bash

# Output command before executing
set -x

# Exit on error
set -e

# Source environment variables of the jenkins slave
# that might interest this worker.
function load_jenkins_vars() {
  if [[ -e "jenkins-env" ]]; then
    cat jenkins-env \
      | grep -E "(DEVSHIFT_TAG_LEN|QUAY_USERNAME|QUAY_PASSWORD|JENKINS_URL|GIT_BRANCH|GIT_COMMIT|BUILD_NUMBER|ghprbSourceBranch|ghprbActualCommit|BUILD_URL|ghprbPullId)=" \
      | sed 's/^/export /g' \
      > ~/.jenkins-env
    source ~/.jenkins-env
  fi
}

function install_deps() {
  # We need to disable selinux for now, XXX
  /usr/sbin/setenforce 0 || :

  # Get all the deps in
  yum -y install \
    docker \
    make \
    git \
    curl

  service docker start

  echo 'CICO: Dependencies installed'
}

function prepare() {
  # Let's test
  make docker-start
  make docker-check-go-format
  make docker-deps
  make docker-build
  echo 'CICO: Preparation complete'
}

function run_tests_without_coverage() {
  make docker-test-unit
  echo "CICO: ran tests without coverage"
}


function run_tests_with_coverage() {
  # Run the unit tests that generate coverage information
  make docker-test-unit-with-coverage

  # Output coverage
  make docker-coverage-all

  # Upload coverage to codecov.io
  cp tmp/coverage.mode* coverage.txt
  bash <(curl -s https://codecov.io/bash) -X search -f coverage.txt -t 6539a1b6-70f9-4ba1-8ab5-d1c8a104ec83 #-X fix

  echo "CICO: ran tests and uploaded coverage"
}

function tag_push() {
  local tag

  tag=$1
  docker tag gitservice-deploy $tag
  docker push $tag
}

function deploy() {
  # Login first
  REGISTRY="quay.io"

  if [[ -n "${QUAY_USERNAME}" && -n "${QUAY_PASSWORD}" ]]; then
    docker login -u ${QUAY_USERNAME} -p ${QUAY_PASSWORD} ${REGISTRY}
  else
    echo "Could not login, missing credentials for the registry"
  fi

  # Build gitservice-deploy
  make docker-image-deploy

  TAG=$(echo $GIT_COMMIT | cut -c1-${DEVSHIFT_TAG_LEN})


  if [[ "$TARGET" = "rhel" ]]; then
    tag_push ${REGISTRY}/openshiftio/rhel-gitservice:$TAG
    tag_push ${REGISTRY}/openshiftio/rhel-gitservice:latest
  else
    tag_push ${REGISTRY}/openshiftio/gitservice:$TAG
    tag_push ${REGISTRY}/openshiftio/gitservice:latest
  fi

  echo 'CICO: Image pushed, ready to update deployed app'
}

function cico_setup() {
  load_jenkins_vars;
  install_deps;
  prepare;
}
