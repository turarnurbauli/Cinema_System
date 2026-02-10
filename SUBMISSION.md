# Assignment 3 Submission Instructions

## Preparation for Submission

### 1. Project Verification

Make sure the project compiles and runs:
```bash
go run .
```

The project should start and display "Cinema System - Assignment 3".

### 2. Creating ZIP Archive

Create a ZIP archive named `Assignment3_Se-2425.zip` (or `Assignment3_TeamName.zip`):

```bash
# On macOS/Linux
cd "/Users/almaslhan/Desktop/AP 3 assignment"
zip -r Assignment3_Se-2425.zip . -x "*.git*" -x "*.DS_Store"

# Or use Finder/File Manager to create archive
```

**Git в архиве (важно для защиты):**  
На защите нужно показать ветки и коммиты от обоих участников. Возможны два варианта:

- **Вариант A:** Включить папку `.git` в ZIP (архив будет больше, но на защите можно сразу показать `git branch -a` и `git log` из распакованной папки).
- **Вариант B:** Не включать `.git` в ZIP, но заранее залить проект на GitHub и на защите открыть репозиторий в браузере и показать ветки и историю коммитов. Репозиторий: **[https://github.com/turarnurbauli/ADP-3](https://github.com/turarnurbauli/ADP-3)**.

```bash
# Вариант A: архив С .git (для показа на защите из папки)
zip -r Assignment3_Se-2425.zip . -x "*.DS_Store"

# Вариант B: архив БЕЗ .git (если покажете репо с GitHub)
zip -r Assignment3_Se-2425.zip . -x "*.git*" -x "*.DS_Store"
```

### 3. Archive Content Verification

The archive should contain:
- ✅ `go.mod` and `go.sum` (if exists)
- ✅ `main.go`
- ✅ `README.md`
- ✅ `docs/` directory with all documentation (including `04_Diagrams_Mermaid.md`)
- ✅ `.gitignore`
- ✅ (по желанию) `.git/` — если выбрали Вариант A

### 4. Structure for Verification

```
Assignment3_Se-2425.zip
├── go.mod
├── main.go
├── README.md
├── .gitignore
├── SUBMISSION.md
├── DEFENSE.md
├── GIT_REPOSITORY_INFO.md
└── docs/
    ├── 01_Project_Proposal.md
    ├── 02_Architecture_Design.md
    ├── 03_Project_Plan.md
    ├── 04_Diagrams_Mermaid.md
    └── README.md
```

## What is Checked During Defense

1. **Project compiles**: `go run .` should work
2. **Git repository**: Branches and commits from both participants should be visible
3. **Documentation**: All three documents should be presented
4. **Architecture**: Diagrams should be clear and match the project

## Git Demonstration Commands

```bash
# Show all branches
git branch -a

# Show commit history
git log --oneline --all --graph

# Show commits by authors
git log --format="%h - %an: %s"
```

## Team Contacts

- **Alkhan Almas** (Se-2425) — https://github.com/AlmasAlkhan
- **Tuar Nurbauli** (Se-2425) — https://github.com/turarnurbauli
