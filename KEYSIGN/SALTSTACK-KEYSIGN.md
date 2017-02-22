## <https://docs.saltstack.com/en/latest/topics/tutorials/multimaster_pki.html#motivation>

### MOTIVATION

The default behaviour of a salt-minion is to connect to a master and accept the masters public key. 
With each publication, the master sends his public-key for the minion to check and if this public-key ever changes, the minion complains and exits. 
Practically this means, that there can only be a single master at any given time.

Would it not be much nicer, if the minion could have any number of masters (1:n) and jump to the next master if its current master died because of a network or hardware failure?

### THE GOAL

Setup the master to sign the public key it sends to the minions and enable the minions to verify this signature for authenticity.

### PREPPING THE MASTER TO SIGN ITS PUBLIC KEY

For signing to work, both master and minion must have the signing and/or verification settings enabled. 
If the master signs the public key but the minion does not verify it, the minion will complain and exit. 
The same happens, when the master does not sign but the minion tries to verify.

The easiest way to have the master sign its public key is to set

```
master_sign_pubkey: True
```

After restarting the salt-master service, the master will automatically generate a new key-pair

```
master_sign.pem
master_sign.pub
```

A custom name can be set for the signing key-pair by setting

```
master_sign_key_name: <name_without_suffix>
```

The master will then generate that key-pair upon restart and use it for creating the public keys signature attached to the auth-reply.

The computation is done for every auth-request of a minion. If many minions auth very often, it is advised to use `conf_master:master_pubkey_signature` and `conf_master:master_use_pubkey_signature` settings described below. If multiple masters are in use and should sign their auth-replies, the signing `key-pair master_sign.*` has to be copied to each master. Otherwise a minion will fail to verify the masters public when connecting to a different master than it did initially. That is because the public keys signature was created with a different signing key-pair.

**CODE**

```python
class MasterKeys(dict):
	def __init__(self, opts):
		self.key = self.__get_keys()
		// customization
		self.pub_signature = fp_.read()

		gen_keys(); 
		self.sign_key = RSA.importKey(f.read())
```

### PREPPING THE MINION TO VERIFY RECEIVED PUBLIC KEYS

The minion must have the public key (and only that one!) available to be able to verify a signature it receives. That public key (defaults to master_sign.pub) must be copied from the master to the minions pki-directory.

Once the verification is successful, the minion can be started in daemon mode again. For the paranoid among us, its also possible to verify the publication whenever it is received from the master. That is, for every single auth-attempt which can be quite frequent. For example just the start of the minion will force the signature to be checked 6 times for various things like auth, mine, highstate, etc.

```python
class AsyncAuth(object):
	self.opts = opts
	self.token = Crypticle.generate_key_string() {
		def generate_key_string(cls, key_size=192):
			key = os.urandom(key_size // 8 + cls.SIG_SIZE)
			return key.encode('base64').replace('\n', '')	
	}
	self.serial = salt.payload.Serial(self.opts)
	self.pub_path = os.path.join(self.opts['pki_dir'], 'minion.pub')
	self.rsa_path = os.path.join(self.opts['pki_dir'], 'minion.pem')

	def __new__(cls, opts, io_loop=None):
		new_auth.__singleton_init__(opts, io_loop=io_loop)
		loop_instance_map[key] = new_auth

	def authenticate(self, callback=None):
		...
		def _authenticate(self):
			...
			def sign_in(self, timeout=60, safe=True, tries=1, channel=None):
			'''
			Send a sign in request to the master, sets the key information and
			returns a dict containing the master publish interface to bind to
			and the decrypted aes key for transport decryption.

			:param int timeout: Number of seconds to wait before timing out the sign-in request
			:param bool safe: If True, do not raise an exception on timeout. Retry instead.
			:param int tries: The number of times to try to authenticate before giving up.

			:raises SaltReqTimeoutError: If the sign-in request has timed out and :param safe: is not set

			:return: Return a string on failure indicating the reason for failure. On success, return a dictionary
			with the publication port and the shared AES key.
			'''
			...

			def minion_sign_in_payload(self):
				'''
				Generates the payload used to authenticate with the master
				server. This payload consists of the passed in id_ and the ssh
				public key to encrypt the AES key sent back from the master.

				:return: Payload dictionary
				:rtype: dict
				'''
				payload = {}
				payload['cmd'] = '_auth'
				payload['id'] = self.opts['id']
				try:
					pubkey_path = os.path.join(self.opts['pki_dir'], self.mpub)
					with salt.utils.fopen(pubkey_path) as f:
						pub = RSA.importKey(f.read())
					cipher = PKCS1_OAEP.new(pub)
					payload['token'] = cipher.encrypt(self.token)
				except Exception:
					pass
				with salt.utils.fopen(self.pub_path) as f:
					payload['pub'] = f.read()
				return payload
		   
			payload = yield channel.send(
				sign_in_payload,
				tries=tries,
				timeout=timeout
			)

			auth['aes'] = self.verify_master(payload, master_pub='token' in sign_in_payload) {

				def verify_master(self, payload, master_pub=True):
				'''
				Verify that the master is the same one that was previously accepted.

				:param dict payload: The incoming payload. This is a dictionary which may have the following keys:
					'aes': The shared AES key
					'enc': The format of the message. ('clear', 'pub', etc)
					'publish_port': The TCP port which published the message
					'token': The encrypted token used to verify the message.
					'pub_key': The RSA public key of the sender.
				:param bool master_pub: Operate as if minion had no master pubkey when it sent auth request, i.e. don't verify
				the minion signature
				:rtype: str
				:return: An empty string on verification failure. On success, the decrypted AES message in the payload.
				'''
					def verify_signing_master(self, payload):
						def verify_pubkey_sig(payload['pub_key'],payload['pub_sig']):
							def verify_signature(path,message, binascii.a2b_base64(sig))
			}
```

## <https://docs.saltstack.com/en/latest/topics/development/architecture.html#id2>




