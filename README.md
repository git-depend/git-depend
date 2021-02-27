# git-depend
Primarily solving the branch merging synchronisation problem.

We want a tool that allows multiple pull requests over multiple repositories to 
be coupled together such that merging one can merge them all atomically.

It should be stateless and installed as an extension to git. It should not be a
server. It should meet XA Compliancy. It should be simple.

## Commands
- git-dep                           : alias for git-depend
- git-dep add project <project-url> : add a project to the dependency
- git-dep rm project                : remove a project
- git-dep merge branch-name         : merge projects into a branch
- git-dep rollback <sha>            : rollback a transaction

git-dep makes the assumption that branch names are all the same. If there is for
example a branch my/feature, then this branch exists in all added git-dep repos
for all projects. There is no reason that an extension in the form of a config
file could not be written in the future to account for various branch names.

## Why not Zuul/Repo/Gitman?
Zuul is a complex tool which requires a lot of overhead and infrastructure.
Repo and GitMan attempt to solve the problem with a reliance on git submodules
which we feel adds to the complexity.

## PoC
- Take a list of repositories.
- Check if the branches that you wish to merge exists.
- Use git-notes to sync the status across multiple repositories and merge.


----

# Problem Statement
- repo B C  
- branch B(1) and C(1)
- You want to only move forward C to C' with C(1) if you can also move B to B' with B(1) merged

## Assumptions
- git updates are atomic
- repo B and C are visible to the automation (though not necessarily to all users)

## Happy Path
- attempt to merge C(1) into C so moving it to C'
- ask C for dependencies (look for a git-note called depends)
   - it returns B (perhaps a URL or a URL and git tag)
- ask B for dependencies (look for a git-note called depends)
   - it returns C (perhaps a URL or URL and git tag)
- put a lock on C (perhaps a git-note called depends-lock with the current SHA of C)
- put a lock on B (perhaps a git-note called depends-lock with the current SHA of B)
   -     [Walk the tree of dependences from bottom to top butting locks in place]
- do git merge of B(1) to B which succeeds assuming the diff merge executes
   - [because this is a git merge, it will call the pre-commit hook which could kick off any CI process]
   - put the success B(1) SHA in the lock on B as being successful
- do git merge of C(1) to C which succeeds assuming diff merge executes
   - [because this is a git merge, it will call the pre-commit hook which could kick off any CI process]
   - put the success of C(1) SHA in the lock on C as being successful
- now check the lock on C to confirm the SHA of C hasn't moved to C'
   - complete the merge
- now check the lock on B to confirm the SHA of B hasn't moved to B'
   - complete the merge
- remove the lock on B
- remove the lock on C

## Problems
- lock placed on repo C or B and by the time the lock is to be undone C is C' or B is B'  
  In effect you retry the lock and merge process.   
  [probably need some exponential limited back-off retry strategy]

- C(1) fails to merge so you need to undo B(1)  
  You have a C lock which has no C(1) SHA in  
  the git merge on C processes can look at the dependency list in the lock and remove the B(1) share from B lock  

- B(1) fails to merge so you either you never need to do C(1)  as above

- corruption of the git-note with the depends in.. [Any ideas]
- corruption of the git-node with locks in. Delete them and end the git-depend

# Quick demo
Let's populate the config file.
```
go run git-dep.go config --author="Finn Ball" --email="finn.ball@codificasolutions.com"
```

Add a project to our dependency:
```
go run git-dep.go add https://github.com/git-depend/repoA.git:my/branch
```

Now commit the note:
```
go run git-dep.go commit
```

This will download the project into the cache, add a note to it and then read the note back.

You can find the projects in the cache:

```
cd ~/.cache/git-depend/42a8ff2939cac2bc934488d5bec881c2e759f7b1/
git notes --ref=git-depend show
```
