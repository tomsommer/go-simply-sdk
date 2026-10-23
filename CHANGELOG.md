# Changelog

## Unreleased

- Deprecated `AuthBasic` and `Config.Account`. Bearer auth with the API key alone is the supported mode; Simply.com identifies the account from the key. Basic auth keeps working for backward compatibility.
- Marked the deprecated `/2/ddns/` and `/2/dyndns/` operations as not planned. DDNS clients should use `https://ddns.simply.com/nic/update`.

- Added the standalone Simply client foundation with Bearer/Basic auth and typed response errors.
- Added product listing and DNS zone/record operations with offline contract tests.
- Added an explicit support matrix for the pinned API 2.7.0 operation inventory.
