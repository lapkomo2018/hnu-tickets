import hashlib

password = input("Введіть пароль: ")

password_hash = hashlib.sha256(
    password.encode("utf-8")
).hexdigest()

if password_hash == "2ed06766795d58a4f22d511a672f20a6b096d3fe5b56af3a744678a9a356fd82":
    print("Доступ дозволено")
else:
    print("Доступ заборонено")
