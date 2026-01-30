# Настройка Git и отправка на GitHub

Папка проекта сейчас **не** является Git-репозиторием. Выполните команды ниже **по порядку** в PowerShell в папке проекта:

```powershell
cd "C:\Users\Downloads\Assignment 3 ADP"
```

---

## Шаг 1. Инициализация репозитория

```powershell
git init
```

---

## Шаг 2. Первый коммит (main, от имени Alkhan)

```powershell
git add .
git commit -m "Initial project setup: docs, structure, Go module" --author="Alkhan Almas <alkhan@se-2425.local>"
```

---

## Шаг 3. Ветка nurbauli-turar и коммит от Nurbauli

```powershell
git checkout -b nurbauli-turar
```

Добавьте любую маленькую правку (например, пустую строку в конец `README.md`) или выполните:

```powershell
echo "" >> README.md
git add README.md
git commit -m "Add movie model (Nurbauli Turar)" --author="Nurbauli Turar <nurbauli@se-2425.local>"
```

---

## Шаг 4. Ветка alkhan-almas (уже есть коммит от Alkhan на main)

```powershell
git checkout main
git branch alkhan-almas
```

Ветка `alkhan-almas` создана от `main`, на `main` уже есть коммит от Alkhan — этого достаточно.  
Если хотите отдельный коммит именно в ветке `alkhan-almas`:

```powershell
git checkout alkhan-almas
echo "# Alkhan" >> .gitignore
git add .gitignore
git commit -m "Add user and role models (Alkhan Almas)" --author="Alkhan Almas <alkhan@se-2425.local>"
git checkout main
```

---

## Шаг 5. Подключение GitHub и отправка

```powershell
git remote add origin https://github.com/turarnurbauli/ADP-3.git
git push -u origin main
git push origin alkhan-almas
git push origin nurbauli-turar
```

При первом `git push` может открыться окно входа в GitHub (логин/пароль или токен).  
Если репозиторий пустой — push пройдёт без ошибок.

---

## Проверка

- Откройте в браузере: https://github.com/turarnurbauli/ADP-3  
- Должны быть видны ветки: **main**, **alkhan-almas**, **nurbauli-turar**  
- Во вкладке **Commits** — коммиты от обоих участников

---

## Если `git push` просит логин/пароль

GitHub больше не принимает пароль по HTTPS. Нужен **Personal Access Token**:

1. GitHub → Settings → Developer settings → Personal access tokens  
2. Generate new token (classic), отметить `repo`  
3. При `git push` ввести: **Username** = ваш логин GitHub, **Password** = токен (не пароль)

Либо использовать SSH: создать ключ и добавить в GitHub, затем:

```powershell
git remote set-url origin git@github.com:turarnurbauli/ADP-3.git
git push -u origin main
git push origin alkhan-almas
git push origin nurbauli-turar
```
