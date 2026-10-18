# Полная инструкция: запуск, сдача и защита

## 1. Что ты сдаёшь

Это индивидуальный SIS по Golang, а не командная презентация. Источники: ссылки в README и REQUIREMENTS.md. В спецификации CP1 оценивается в 5 баллов. Руководство распределяет до 2.5 за реализацию и до 2.5 за защиту; без обязательной защиты — 0. Защита обычно около 2 минут, без презентации и слайдов. По сообщению задания защита проходит на практике Week 8.

Приложение Text Analyzer получает текст и создаёт задачу на его анализ. Задача — единица работы программы. Это не список дел. Результат — количество слов, Unicode-символов, букв и цифр.

## 2. Что сделать до дедлайна

1. Распакуй архив. Открой папку sis-text-analyzer в VS Code. В корне должны быть go.mod, README.md, task и cmd.
2. Установи Go 1.27 с https://go.dev/dl/ . После установки перезапусти терминал. Проверь `go version`: требуется go1.27.x. Если другая версия, не меняй go.mod ради обхода требования.
3. В терминале в этой папке выполни команды ниже.
4. Прочитай task/task.go, запусти demo, изучи тесты и объяснения ниже. Измени хотя бы текст демо самостоятельно и предскажи результат до запуска.
5. Создай свой публичный GitHub-репозиторий и загрузи содержимое папки. Проверь страницу без входа в аккаунт.
6. Отправь ссылку в Teams до 22:00 сегодня, как написано в твоём задании. Руководство также просит ссылку на соответствующий commit/link/tag — добавь ссылку на конкретный commit.
7. Открой свою отправку Teams и убедись, что ссылка сохранилась. Сам архив вместо ссылки на GitHub не выполняет указанное требование.

Команды запуска (PowerShell, cmd, терминал VS Code):

```sh
go version
go run ./cmd/demo
gofmt -l .
go vet ./...
go test -v ./...
go build ./...
```

`go run` запускает демо; `go test` проверяет поведение; `go vet` ищет подозрительные конструкции; `gofmt -l` перечисляет неформатированные файлы. Если список не пуст, выполни `gofmt -w .` и повтори проверки. `go build ./...` компилирует пакеты. Результаты фактической проверки здесь лежат в VALIDATION.txt.

Если `go` не найден — проверь установку и PATH, перезапусти терминал. Если терминал сообщает, что go.mod не найден, перейди в корень распакованного проекта. Не запускай отдельный main.go из другой папки; используй `go run ./cmd/demo` из корня.

## 3. Как загрузить на GitHub

Вариант через сайт: GitHub → New repository → имя sis-text-analyzer → Public → Create repository. Затем Add file → Upload files. Загрузи содержимое распакованной папки с сохранением task, cmd и docs. Убедись, что в корне репозитория виден go.mod, а README отображается на главной странице. Нажми Commit changes. Не загружай только ZIP.

Вариант через Git (после создания пустого public repo; подставь свой логин в URL):

```sh
git init
git add .
git commit -m "Implement SIS checkpoint 1 text analyzer"
git branch -M main
git remote add origin https://github.com/YOUR_USERNAME/sis-text-analyzer.git
git push -u origin main
```

YOUR_USERNAME — шаблон, его нужно заменить. При запросе GitHub пройди обычную авторизацию; пароль аккаунта не вставляй в исходники. На странице репозитория открой последний commit и скопируй его URL. Отправь в Teams две реальные ссылки: repository и checkpoint commit. Я не создал репозиторий в твоём аккаунте и не отправлял задание за тебя.

## 4. Как работает код, по порядку

### go.mod

`module sis-text-analyzer` задаёт путь модуля. Поэтому импорт своего пакета в main — `sis-text-analyzer/task`. `go 1.27.0` задаёт требуемую версию языка/инструментария. Сторонних библиотек нет, поэтому go.sum не нужен.

### Типы в task/task.go

`Status` — отдельный строковый тип. Константы Pending, Processing, Completed уменьшают риск случайных строк в коде. Task хранит ID int, Input string, Status и Result. Result — отдельная структура с четырьмя счётчиками.

`ErrInvalidInput`, `ErrNotFound`, `ErrInvalidState` — заранее объявленные ошибки. В Get и Process к ошибке добавляется ID через fmt.Errorf и `%w`, сохраняя возможность `errors.Is(err, ErrNotFound)`.

### Интерфейс Analyzer

Интерфейс описывает поведение: `Analyze(string) Result`. TextAnalyzer реализует его автоматически, потому что имеет метод с нужной сигнатурой. Ключевого слова implements в Go нет. Service вызывает интерфейс, а не жёстко привязанный вызов TextAnalyzer. Тест подменяет анализатор наблюдателем, который проверяет промежуточное состояние.

### TextAnalyzer.Analyze

`strings.Fields` делит строку по пробельным символам. `len` получившегося среза — число слов. Пунктуация отдельно не удаляется. `utf8.RuneCountInString` считает руны; цикл `range` по строке тоже перебирает руны. unicode.IsLetter и unicode.IsDigit определяют буквы и цифры.

Пример `Hello Go world`: 3 слова; 14 символов = 5+1+2+1+5; 12 букв = 5+2+5; 0 цифр. Пример `Сәлем Go 2026`: 3 слова; 13 символов = 5+1+2+1+4; 7 букв; 4 цифры. Эти примеры проверяются тестами.

Ограничение определения: руна — Unicode code point, а не обязательно один визуальный символ. Сочетания с диакритикой и составные emoji могут содержать несколько рун. `len(text)` считал бы байты, что даёт другое значение для кириллицы.

### Service и конструктор

`map[int]Task` связывает ID со значением задачи. nextID начинается с 1 и увеличивается только после успешной валидации. NewService вызывает make для карты и задаёт TextAnalyzer. Создавать сервис нужно конструктором: нулевой Service не готов к работе.

### Create

Проверяет UTF-8 и непустой после TrimSpace текст. При ошибке возвращает Task{} и error, ничего не записывая. Иначе создаёт Pending, записывает в карту и увеличивает счётчик. Оригинальный текст не обрезается: пробелы учитываются в результате.

### Get

`t, ok := s.tasks[id]` — получение из map с проверкой существования ключа. Если !ok, возвращается ErrNotFound с контекстом. Иначе возвращается копия Task. Если изменить её снаружи, задача в карте останется прежней. В этой модели нет срезов, карт или указателей внутри Task, которые могли бы оставить общий изменяемый объект.

### Process

1. Получает задачу через Get, возвращает ошибку для неизвестного ID.
2. Проверяет Pending; для других состояний возвращает ErrInvalidState.
3. Меняет локальную копию на Processing и записывает её в карту.
4. Вызывает Analyze для входного текста.
5. Сохраняет Result, ставит Completed и снова записывает в карту.
6. Возвращает завершённую задачу.

Важно: map содержит значения, поэтому изменения локальной копии нужно явно записать обратно. Обработка синхронная: вызывающий код ждёт завершения Process. Промежуточное Processing существует внутри вызова; тест проверяет его анализатором-наблюдателем. Встроенная обработка корректного текста не имеет ожидаемой ошибки вычисления, поэтому состояния Failed здесь нет.

### main

Создаёт сервис, последовательно анализирует два текста, печатает результат. Затем демонстрирует три ожидаемые ошибки: пробельный ввод, отсутствующий ID и повторную обработку. Это демонстрация, не интерактивное меню. Неожиданная ошибка run передаётся main, печатается в stderr и завершает процесс с кодом 1.

## 5. Что проверяют тесты

- TestCreateAndGet: поля новой задачи, Pending, поиск и защита хранения от изменения возвращённой копии.
- TestInvalidInput: пустой текст, пробелы, табы/переносы, Unicode-пробел, некорректный UTF-8; отсутствие записи и расходования ID при отказе.
- TestMissingTask: Get и Process для неизвестных ID.
- TestProcessResults: таблица входов и точных ожидаемых счётчиков, Completed, сохранение результата и отказ повторной обработки.
- TestIndependentTasks: разные ID, результаты не смешиваются, обработка одной не меняет другую.
- TestProcessingState: состояние действительно Processing во время вычисления, повторный Process в этот момент отклоняется.

Table-driven test — тест, где набор случаев хранится в таблице, а одна логика проверки проходит по каждому через t.Run. Здесь это TestProcessResults. Проверка результата — сравнение структуры с ожидаемым значением, а не просто запуск без падения.

## 6. Готовый рассказ на защите

Русский:

«Мой проект — синхронный анализатор текста. Задача содержит уникальный в рамках сервиса ID, исходный текст, состояние и результат. Create проверяет UTF-8 и отклоняет пустой или пробельный ввод, затем сохраняет Pending в map. Get возвращает копию задачи или ошибку, если ID отсутствует. Process разрешён только для Pending. Он сохраняет Processing, вычисляет статистику через интерфейс Analyzer, затем сохраняет результат и Completed. Я считаю слова через strings.Fields, символы как Unicode-руны, а буквы и цифры через unicode. Методы сервиса используют pointer receiver, потому что изменяют его состояние. Анализатор использует value receiver, потому что у него нет изменяемых полей. Тесты проверяют результаты, ошибки, состояния и независимость задач. Хранение только в памяти; concurrency и HTTP на этом этапе не реализованы».

English (simple):

“My project is a synchronous text analyzer. Each task has an ID, input text, a status, and a result. Create validates the text and stores a pending task in a map. Get returns a task copy or an error. Process only accepts pending tasks. It sets processing, calls the Analyzer interface, saves the result, and sets completed. The result contains words, Unicode characters, letters, and digits. Service methods use pointer receivers to update service state. The analyzer uses a value receiver because it has no mutable fields. My tests check results, errors, states, and independent tasks. Data is stored only in memory.”

Не заучивай без понимания: преподаватель может выбрать любой участок. Покажи код и объясни путь конкретного текста от Create до результата.

## 7. Возможные вопросы — короткие ответы RU / EN

**Что такое struct?** Группа именованных полей одного типа объекта; здесь Task. / A struct groups named fields; Task is the task model.

**Метод и функция?** У метода есть receiver, например `(s *Service)`; обычная функция не привязана к receiver. / A method has a receiver; a function does not.

**Зачем *Service?** Чтобы изменять исходный nextID и состояние сервиса. Карта сама ссылается на внутреннее хранилище, но value receiver копировал бы nextID и мог привести к повторению ID. / It updates the original service, including nextID.

**Зачем value receiver у TextAnalyzer?** Он не хранит изменяемого состояния. / It has no mutable state.

**Что такое interface?** Набор требуемых методов; позволяет заменить реализацию. / An interface defines required methods and allows different implementations.

**Зачем map, а не slice?** Нужен поиск по ID; map напрямую связывает ID с задачей. Со slice потребовался бы поиск или отдельное соглашение об индексах. / A map directly associates an ID with its task.

**Что такое slice здесь?** strings.Fields возвращает []string — срез слов; его длина — число слов. / Fields returns a slice of words.

**Почему нельзя менять s.tasks[id].Status напрямую?** Элемент map со значением struct не адресуем как обычная переменная. Получаем копию, меняем и присваиваем обратно. / Read the struct value, update it, and store it back.

**Что возвращает Get при неизвестном ID?** Task{} и обёрнутую ErrNotFound. / A zero Task and a wrapped ErrNotFound.

**Почему error, а не panic?** Плохой ввод и неизвестный ID — ожидаемые ошибки приложения; вызывающий код должен их обработать. / They are expected application errors.

**Что делает %w?** Оборачивает error, сохраняя возможность errors.Is. / It wraps an error and preserves errors.Is checks.

**Можно ли обработать дважды?** Нет, после первого раза Completed, а разрешён только Pending. / No. Process only accepts pending tasks.

**Почему Processing, если нет goroutine?** Это реальный этап внутри синхронного Process; тест проверяет его во время вызова Analyzer. / Processing exists during the synchronous calculation.

**Что после перезапуска?** Данные исчезают, ID снова начинаются с 1. / Data is lost and IDs restart.

**Потокобезопасно?** Нет. Карта и счётчик без синхронизации; для CP1 вызовы последовательные. / No. It is designed for sequential use in checkpoint 1.

**Сложность?** Анализ проходит по тексту линейно относительно его длины. Fields дополнительно создаёт срез слов. Хранение всех задач требует память для всех сохранённых текстов и записей. / Analysis is linear in text length; Fields allocates a word slice.

**Почему нет БД/HTTP?** CP1 требует in-memory synchronous приложение; расширения относятся к следующим checkpoint. / They belong to later checkpoints.

**Что если изменить результат Get?** Меняется копия; оригинал в map остаётся прежним. / Only the returned copy changes.

## 8. Самопроверка без подсказок

1. Для `Go 123!` предскажи результат: words=2, characters=7, letters=2, digits=3. Проверь вручную: G,o,пробел,1,2,3,!.
2. Объясни каждый return в Create, Get и Process.
3. Найди в тестах проверку Completed и повторной обработки.
4. Объясни, зачем запись в map выполняется дважды в Process.
5. Замени один вход демо и добавь собственный случай в таблицу тестов.
6. Закрой инструкцию и расскажи проект на английском или русском за 1–2 минуты.

Смысл подготовки — уметь самому прочитать и изменить код. Готовый архив сам по себе не гарантирует защиту или оценку.

## 9. Проверяемые источники

Требования и правила защиты: ссылки на закреплённый commit преподавателя в README.md и REQUIREMENTS.md. Поведение конкретного проекта: task/task.go, task/task_test.go и фактический вывод VALIDATION.txt.

Документация стандартной библиотеки для проверки терминов:
- https://go.dev/ref/spec — types, methods, interfaces, map indexing, range.
- https://pkg.go.dev/strings#Fields
- https://pkg.go.dev/unicode/utf8#RuneCountInString
- https://pkg.go.dev/unicode#IsLetter
- https://pkg.go.dev/unicode#IsDigit
- https://pkg.go.dev/errors#Is
- https://pkg.go.dev/fmt#Errorf
- https://pkg.go.dev/testing
