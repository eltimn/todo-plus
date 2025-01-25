# todo-plus

## Note

If you run `go mod tidy` be sure to also run `templ generate` as tidy will remove the entry for a-h/templ. And then run `gomod2nix`. See `tidy` task in the Taskfile.

## Todo

- create an interface for session
- finish implementing repo pattern ???
- add a sql builder
  - [go-sqlbuilder](https://github.com/huandu/go-sqlbuilder)
  - [sqlc](https://sqlc.dev) ***
  - [jet](https://github.com/go-jet/jet)
- github action
	- compile/test
- deploy
	- Look into NixOps
	- Deploy using pulumi/terraform cdk to aws ec2 instances
- Add lots more tests
- try to use dockerTools (probably requires being able to build both packages at the same time)

## Resources

- [HTMX Book](https://hypermedia.systems/book/foreword/)
- [.gitxxx files](https://www.dudley.codes/posts/2020.02.16-git-lost-in-translation/)
- [automatically-resolve-npm-package-lock-conflicts-using-git](https://martin.beryllium.net/2019/08/02/automatically-resolve-npm-package-lock-conflicts-using-git/)
- [Git Attributes Blog Post](https://pablorsk.medium.com/be-a-git-ninja-the-gitattributes-file-e58c07c9e915)
- https://www.dudley.codes/posts/2020.05.19-golang-structure-web-servers/

