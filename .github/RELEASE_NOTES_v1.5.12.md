# Release Notes: v1.5.12

This release adds the requested client creation defaults, direct expiry date entry, and certificate/domain setup improvements.

## Client management

- New clients default to a timestamp name, 50 GiB traffic, one month of validity, and an IP limit of three.
- Suggested validity changes to three, six, or twelve months at 150, 300, or 600 GiB.
- Expiry dates accept direct input such as `26.11.6`.

## Installation and TLS

- Fresh installs use random panel and subscription ports and prompt for custom admin credentials.
- Existing installations keep their current settings and skip setup prompts.
- Fresh installs can select an existing certificate under `/root/cert` and bind its domain.
- Settings can fill panel and subscription certificate paths for the selected domain.
