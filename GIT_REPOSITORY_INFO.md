# Git Repository Information

## GitHub Repository

**URL:** [https://github.com/turarnurbauli/ADP-3](https://github.com/turarnurbauli/ADP-3)  
**Clone:** `git clone https://github.com/turarnurbauli/ADP-3.git`

Чтобы отправить локальный проект на GitHub (если репозиторий на GitHub ещё пустой):

```bash
git remote add origin https://github.com/turarnurbauli/ADP-3.git
git push -u origin main
git push origin alkhan-almas
git push origin nurbauli-turar
```

## Repository Setup Status ✅

This project has a Git repository configured according to Assignment 3 requirements.

## What is a Git Repository?

A **Git repository** is a version control system that tracks all changes to your project files. It allows you to:
- See the history of all changes
- Work on different features in separate branches
- Collaborate with team members
- Track who made what changes and when

## Current Repository Status

### Branches (Ветки)

The repository has 3 branches:

1. **`main`** - Main branch with all merged code
2. **`alkhan-almas`** - Branch for Alkhan Almas's work
3. **`nurbauli-turar`** - Branch for Nurbauli Turar's work

### Commits (Коммиты)

Total commits: **6**

#### Commit History:

```
* 545b57e (HEAD -> main) Translate all documentation to English
* 17c0bf0 Simplify project: keep only documentation and minimal code
* 9f5ab03 Add documentation index
* a5eb75c (nurbauli-turar) Add movie model (Nurbauli Turar)
* 32b874f (alkhan-almas) Add user and role models (Alkhan Almas)
* d2f7581 Initial project setup: base structure and configuration files
```

### Commits by Team Member:

**Alkhan Almas:**
- Initial project setup
- User and role models
- Architecture design
- Documentation translation

**Nurbauli Turar:**
- Movie model (committed in nurbauli-turar branch)

## How to Verify Repository

### View all branches:
```bash
git branch -a
```

### View commit history:
```bash
git log --oneline --all --graph
```

### View commits by author:
```bash
git log --format="%h - %an: %s" --all
```

### View commits in specific branch:
```bash
git log alkhan-almas --oneline
git log nurbauli-turar --oneline
```

## Assignment 3 Requirements Met ✅

- ✅ Git repository initialized
- ✅ Branches created for each team member
- ✅ Commits from both team members
- ✅ All branches merged into main
- ✅ Complete commit history visible

## For Defense

During the defense, you may be asked to:
1. Show the repository structure: `git branch -a`
2. Show commit history: `git log --oneline --all --graph`
3. Explain the branching strategy
4. Show commits from each team member

**Если показываете с GitHub:** перед защитой выполните `git push -u origin main` и `git push origin alkhan-almas` / `git push origin nurbauli-turar`, чтобы на [github.com/turarnurbauli/ADP-3](https://github.com/turarnurbauli/ADP-3) были видны все ветки и коммиты.
