# go-ecs

This is an archetype-based ECS framework implement in Golang.

Highly inspired by [flex](https://github.com/SanderMertens/flecs).

## Performance

The goal is to maximize performance, no `reflect` everywhere, but also using a little to make life better.

## Project Status

We are waiting Golang support "Generic Methods".  
See: <https://github.com/golang/go/issues/49085>

> 一些个人看法：当前Golang语言的开发谷歌是众所周知的双标，社区提案基本很少通过，内部提的proposal却可以一路绿灯（甚至是那些社区提出过无数遍被否决过的）。
> 所以我认为Go短期内不可能支持泛型方法，除非某个谷歌内部员工写代码时需要用到这个特性，我们可以拭目以待。😁😁😁

## Project Structure

The ecs implement is in `/internal/core`, exports its api on `/api.go`.
It's because I want to keep private fields private, but allow `/reflect` package can access them.
