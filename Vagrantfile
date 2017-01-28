# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "bento/ubuntu-16.04"

  config.vm.network "forwarded_port", guest: 3000, host: 3000
  config.vm.synced_folder '.', '/home/vagrant/wdgo/src/approvy'

  config.vm.provision "shell", path: "setup/vagrant_provision.sh", privileged: false

  config.vm.provider "virtualbox" do |v|
    v.memory = ENV['VAGRANT_MEMORY'] || 2048
    v.cpus = ENV['VAGRANT_CPUS'] || 2
  end

  config.vm.provision "docker" do |d|
    d.run "redis", image: "redis:3.2.6", args: "-p 6379:6379"
  end
end
