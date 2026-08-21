# go-ecs

This is an archetype-based ECS framework implement in Golang.

Highly inspired by [flex](https://github.com/SanderMertens/flecs).

## Performance

The goal is to maximize performance, no `reflect` everywhere, but also using a little to make life better.

## Project Structure

The ecs implement is in `/internal/core`, exports its api on `/api.go`.
It's because I want to keep private fields private, but allow `/reflect` package can access them.
