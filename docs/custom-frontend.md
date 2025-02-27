# Custom Frontend for XSS Hunter

XSS Hunter supports using custom frontend files for the admin interface. This allows you to provide your own login and admin pages without modifying the codebase.

## Environment Variables

To use custom frontend files, set the following environment variables:

- `CUSTOM_LOGIN_PATH`: Path to your custom login HTML file (default: `./src/login.html`)
- `CUSTOM_ADMIN_PATH`: Path to your custom admin HTML file (default: `./src/admin.html`)

## Docker Example

You can mount your custom frontend files when using Docker:
