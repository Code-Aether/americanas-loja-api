#!/bin/bash

SECRET=

if command -v openssl >/dev/null 2>&1; then
    SECRET=$(openssl rand -base64 64 | tr -d '[:space:]')
fi

if [ -z "$SECRET" ]; then
  if command -v python3 >/dev/null 2>&1; then
      SECRET=$(python3 -c "import secrets; print(secrets.token_urlsafe(64))")
  fi
fi

if [ -z "$SECRET" ]; then 
    echo "Could not generate a new secret"
else
    printf "\"%s\"" "$SECRET"
fi
