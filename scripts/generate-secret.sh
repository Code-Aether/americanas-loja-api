#!/bin/bash

echo "Generating JWT secret"
SECRET=

if command -v openssl >/dev/null 2>&1; then
    SECRET=$(openssl rand -base64 64 | tr -d '\n')
    echo "Generated with OpenSSL:"
    echo "Export like this:"
    echo "export JWT_SECRET=\"$SECRET\""
    echo ""
fi

if [ -z "$SECRET" ]; then
  if command -v python3 >/dev/null 2>&1; then
      SECRET=$(python3 -c "import secrets; print(secrets.token_urlsafe(64))")
      echo "Generated with Python3:"
      echo "Export like this:"
      echo "export JWT_SECRET=\"$SECRET\""
      echo ""
  fi
fi

if [ -z "$SECRET" ]; then 
    echo "Could not generate a new secret"
fi
