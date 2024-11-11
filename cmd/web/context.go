package main

type contextKey string

const isAuthenticatedContextKey = contextKey("isAuthenticated")
const userNameContextKey = contextKey("username")
