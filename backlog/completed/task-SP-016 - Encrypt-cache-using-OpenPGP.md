---
id: SP-016
title: Encrypt cache using OpenPGP
status: Done
assignee: []
created_date: '2025-12-24 17:15'
labels: []
dependencies: []
---

# Description

Encrypt the local cache (specifically the API key and user info) using OpenPGP to improve security.
The implementation should leverage the `gpg` binary to utilize the user's existing GPG Agent and Keyring.
