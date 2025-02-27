# Custom Frontend for XSS Hunter

XSS Hunter supports using a custom frontend instead of the default UI. This guide explains how to set up your own frontend.

## Overview

By default, XSS Hunter serves built-in HTML files from the `./src` directory. With the custom frontend feature, you can:

- Replace the login and admin pages with your own versions
- Serve additional static assets like JavaScript, CSS, and images
- Use modern frontend frameworks like React, Vue, or Svelte

## Setting Up a Custom Frontend

### 1. Configure the Environment Variable

Set the `CUSTOM_FRONTEND_DIR` environment variable to point to your frontend directory:
