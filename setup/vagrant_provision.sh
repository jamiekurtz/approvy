#!/bin/bash

# Original script on https://github.com/orendon/vagrant-rails

export DEBIAN_FRONTEND=noninteractive

# enable console colors
sed -i '1iforce_color_prompt=yes' ~/.bashrc
. ~/.bashrc

# basic packages
sudo apt-get -y update
sudo apt-get -y install vim wget curl
sudo apt-get -y install ntp

# C compiler for Ubuntu
sudo apt-get -y install build-essential

# install golang tools
wget -nv https://storage.googleapis.com/golang/go1.7.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.7.4.linux-amd64.tar.gz

# install some dev tools
sudo apt-get -y install redis-tools sqlite3

# cleanup
sudo apt-get clean

# configure golang workspace
sudo chown -R vagrant:vagrant /home/vagrant/wdgo


# create some content for prepending to .bashrc
> /tmp/bashrc-content
echo 'export GOPATH=$HOME/wdgo' >> /tmp/bashrc-content
echo 'export APPROVYPATH=$GOPATH/src/approvy' >> /tmp/bashrc-content
echo 'export PATH=$PATH:/usr/local/go/bin' >> /tmp/bashrc-content
echo 'export PATH=$PATH:/$GOPATH/bin' >> /tmp/bashrc-content
echo '' >> /tmp/bashrc-content

# backup .bashrc and create the updated version
cat /tmp/bashrc-content | cat - ~/.bashrc > /tmp/bashrc-new
if [ ! -f ~/.bashrc.original ]; then cp  ~/.bashrc ~/.bashrc.original; fi
mv /tmp/bashrc-new ~/.bashrc

