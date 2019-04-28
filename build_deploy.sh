#!/bin/bash
set -e

# This script will build the binary and upload it together with the frontend
# and restart the service.
# NOTE: this does not a full deployment, and assumes config and keys are already provided!
# Dependencies: bash, git, ssh

if [ ! -z "$(git status --untracked-files=no --porcelain)" ]; then
  echo "won't deploy as long as there are uncommitted changes!"
  exit 1
fi

version=$(git rev-parse --short HEAD)
userHost=fsadmin@geofs.uni-muenster.de
targetDir=/srv/UserManager
targetBinary=$targetDir/UserManager_$version

echo "building binary.."
go build

# upload stuff
echo "shooting it into the interweb.."
scp UserManager $userHost:$targetBinary
scp -r public $userHost:$targetDir/public

# set new version & restart service
echo "unleashing your latest creation.."
ssh $userHost 'ln -sf $targetBinary $targetDir/userManager'
ssh $userHost 'systemctl restart userManager'
