import boto3
import base64
from Crypto import Random
from Crypto.Cipher import AES
from Crypto.Util import Padding


def get_blob(name: str, *, with_decryption: bool = False) -> str:
    response = boto3.client("ssm").get_parameter(Name=name, WithDecryption=with_decryption)
    return response["Parameter"]["Value"]

def decrypt_data_key(cipher_text_blob: str) -> str:
    response = boto3.client("kms").decrypt(CiphertextBlob=base64.b64decode(cipher_text_blob))
    return base64.b64encode(response["Plaintext"]).decode("utf-8")

def encrypt(plain_text: str, data_key: str) -> str:
    key = base64.b64decode(data_key)
    iv = Random.get_random_bytes(AES.block_size)
    cipher = AES.new(key, AES.MODE_CBC, iv)
    data = Padding.pad(plain_text.encode("utf-8"), AES.block_size, "pkcs7")
    return base64.b64encode(iv + cipher.encrypt(data)).decode("utf-8")

def decrypt(encrypted_text: str, data_key: str) -> str:
    key = base64.b64decode(data_key)
    encrypted_text = base64.b64decode(encrypted_text)
    iv = encrypted_text[:AES.block_size]
    cipher = AES.new(key, AES.MODE_CBC, iv)
    data = Padding.unpad(cipher.decrypt(encrypted_text[AES.block_size:]), AES.block_size, "pkcs7")
    return data.decode("utf-8")
