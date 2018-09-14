python acme_tiny.py --account-key ./account.key --csr ./domain.etym.csr --acme-dir ../challenges/ > ./signed.etym.crt
wget -O - https://letsencrypt.org/certs/lets-encrypt-x3-cross-signed.pem > intermediate.pem
signed.etym.crt intermediate.pem > chaind.pem
wget -O - https://letsencrypt.org/certs/isrgrootx1.pem > root.pem
cat intermediate.pem root.pem > full_chained.pem
