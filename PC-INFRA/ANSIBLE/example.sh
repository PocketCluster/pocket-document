
ansible all --inventory-file=./ansible_hosts -m ping

ansible us-east-01 -i ansible_hosts -m playbook/essential/essential.yaml

ansible all -i ansible_hosts -s -m shell -a 'apt-get install -y --no-install-suggests nginx'

ansible all -i ansible_hosts -s -m shell -a 'apt-get install -y --no-install-suggests htop'

