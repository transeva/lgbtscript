<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LGBTScript — Документация языка</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #fdf6f9 0%, #f0e6ef 100%);
            color: #2d1b2e;
            line-height: 1.7;
            padding: 20px;
        }
        
        .container {
            max-width: 1100px;
            margin: 0 auto;
            background: rgba(255, 255, 255, 0.92);
            backdrop-filter: blur(10px);
            border-radius: 32px;
            padding: 50px 60px;
            box-shadow: 0 20px 60px rgba(180, 80, 160, 0.15);
            border: 1px solid rgba(255, 182, 193, 0.2);
        }
        
        h1 {
            font-size: 3.2rem;
            font-weight: 800;
            background: linear-gradient(135deg, #e84393, #6c5ce7, #0984e3);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            margin-bottom: 8px;
            letter-spacing: -1px;
        }
        
        .subtitle {
            font-size: 1.2rem;
            color: #6c5b7b;
            margin-bottom: 40px;
            border-bottom: 3px solid #f8e1f0;
            padding-bottom: 20px;
        }
        
        .subtitle span {
            background: linear-gradient(135deg, #fd79a8, #a29bfe);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            font-weight: 600;
        }
        
        h2 {
            font-size: 2rem;
            font-weight: 700;
            color: #4a2c4a;
            margin-top: 45px;
            margin-bottom: 20px;
            padding-left: 16px;
            border-left: 5px solid #e84393;
            display: flex;
            align-items: center;
            gap: 12px;
        }
        
        h2 .emoji {
            font-size: 1.6rem;
        }
        
        h3 {
            font-size: 1.4rem;
            font-weight: 600;
            color: #5a3d5a;
            margin-top: 30px;
            margin-bottom: 12px;
        }
        
        h4 {
            font-size: 1.1rem;
            font-weight: 600;
            color: #6c4b6c;
            margin-top: 22px;
            margin-bottom: 8px;
        }
        
        p {
            margin-bottom: 16px;
            color: #3d2a3e;
        }
        
        ul, ol {
            margin: 12px 0 20px 28px;
            color: #3d2a3e;
        }
        
        li {
            margin-bottom: 6px;
        }
        
        code {
            font-family: 'SF Mono', 'Fira Code', 'JetBrains Mono', monospace;
            background: #f7edf5;
            padding: 2px 10px;
            border-radius: 8px;
            font-size: 0.9rem;
            color: #c0397a;
            border: 1px solid #f0dce8;
        }
        
        pre {
            background: #1e1220;
            color: #f0e6ef;
            padding: 22px 28px;
            border-radius: 16px;
            overflow-x: auto;
            font-family: 'SF Mono', 'Fira Code', 'JetBrains Mono', monospace;
            font-size: 0.92rem;
            line-height: 1.6;
            margin: 16px 0 24px 0;
            border: 1px solid #3d2a3e;
            box-shadow: inset 0 2px 8px rgba(0,0,0,0.3);
        }
        
        pre .comment { color: #8a7a8a; }
        pre .keyword { color: #fd79a8; font-weight: 600; }
        pre .string { color: #a29bfe; }
        pre .number { color: #fdcb6e; }
        pre .function { color: #74b9ff; }
        pre .operator { color: #ffb8c6; }
        pre .type { color: #55efc4; }
        
        .badge {
            display: inline-block;
            background: linear-gradient(135deg, #e84393, #6c5ce7);
            color: white;
            font-size: 0.7rem;
            font-weight: 700;
            padding: 2px 12px;
            border-radius: 20px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-left: 8px;
            vertical-align: middle;
        }
        
        .badge-new {
            background: linear-gradient(135deg, #00b894, #00cec9);
        }
        
        .card {
            background: #faf4f8;
            border-radius: 16px;
            padding: 20px 24px;
            margin: 16px 0;
            border-left: 4px solid #e84393;
        }
        
        .card-example {
            background: #1e1220;
            border-left: 4px solid #fd79a8;
            padding: 18px 22px;
            border-radius: 12px;
            margin: 12px 0;
        }
        
        .card-example code {
            background: transparent;
            color: #f0e6ef;
            border: none;
            padding: 0;
        }
        
        .grid-2 {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
            margin: 16px 0;
        }
        
        @media (max-width: 768px) {
            .container {
                padding: 24px 18px;
                border-radius: 20px;
            }
            h1 {
                font-size: 2.2rem;
            }
            .grid-2 {
                grid-template-columns: 1fr;
            }
            pre {
                padding: 14px 16px;
                font-size: 0.82rem;
            }
        }
        
        .toc {
            background: #faf0f6;
            padding: 20px 28px;
            border-radius: 16px;
            margin: 24px 0 32px 0;
        }
        
        .toc a {
            color: #6c3b6c;
            text-decoration: none;
            border-bottom: 1px dashed #dbb8d0;
        }
        
        .toc a:hover {
            color: #e84393;
            border-bottom-color: #e84393;
        }
        
        .toc ul {
            columns: 2;
            column-gap: 40px;
        }
        
        @media (max-width: 600px) {
            .toc ul {
                columns: 1;
            }
        }
        
        hr {
            border: none;
            height: 2px;
            background: linear-gradient(to right, transparent, #f0dce8, transparent);
            margin: 40px 0;
        }
        
        .footer {
            margin-top: 50px;
            padding-top: 24px;
            border-top: 2px solid #f0dce8;
            text-align: center;
            color: #8a7a8a;
            font-size: 0.95rem;
        }
        
        .footer .rainbow {
            background: linear-gradient(to right, #e84393, #fd79a8, #fdcb6e, #00b894, #0984e3, #6c5ce7, #e84393);
            background-size: 200% auto;
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            animation: rainbow-shift 3s linear infinite;
        }
        
        @keyframes rainbow-shift {
            0% { background-position: 0% center; }
            100% { background-position: 200% center; }
        }
    </style>
</head>
<body>
<div class="container">

    <h1>🌈 LGBTScript</h1>
    <div class="subtitle">
        Инклюзивный, выразительный язык программирования для всех<br>
        <span>«Любовь — это код, который не требует компиляции»</span>
    </div>

    <!-- Оглавление -->
    <div class="toc">
        <strong style="font-size: 1.1rem;">📑 Содержание</strong>
        <ul style="margin-top: 10px;">
            <li><a href="#intro">Введение</a></li>
            <li><a href="#quickstart">Быстрый старт</a></li>
            <li><a href="#types">Типы данных</a></li>
            <li><a href="#variables">Переменные</a></li>
            <li><a href="#operators">Операторы</a></li>
            <li><a href="#arrays">Массивы</a></li>
            <li><a href="#control">Управление потоком</a></li>
            <li><a href="#functions">Функции</a></li>
            <li><a href="#builtins">Встроенные функции</a></li>
            <li><a href="#error">Обработка ошибок</a></li>
            <li><a href="#libraries">Библиотеки</a></li>
            <li><a href="#cli">CLI</a></li>
            <li><a href="#examples">Примеры</a></li>
        </ul>
    </div>

    <!-- Введение -->
    <h2 id="intro"><span class="emoji">🏳️‍🌈</span> Введение</h2>
    <p>
        <strong>LGBTScript</strong> — это интерпретируемый язык программирования, созданный с идеей инклюзивности,
        дружелюбности и выразительности. Ключевые слова вдохновлены тематикой ЛГБТК+ сообщества, что делает код
        не только функциональным, но и символичным.
    </p>
    <p>
        Язык поддерживает: статическую типизацию с указанием типа через ключевые слова, массивы, функции,
        экспорт, обработку ошибок, работу с файлами, HTTP-запросы, регулярные выражения, криптографию и многое другое.
    </p>

    <div class="card">
        <strong>✨ Особенности:</strong>
        <ul style="margin-top: 8px;">
            <li>🔤 4 базовых типа: <code>lesbian</code> (строка), <code>gay</code> (число), <code>trans</code> (float), <code>nonbinary</code> (bool)</li>
            <li>📦 Массивы с динамическим размером</li>
            <li>🧩 Функции с экспортом (<code>export rainbow</code>)</li>
            <li>🔄 Управление потоком: <code>gender</code> (if), <code>queer</code> (else), <code>pride</code> (while)</li>
            <li>📂 Встроенные функции для работы с файлами, HTTP, JSON, криптографией, регулярками и системой</li>
            <li>📚 Подключение библиотек через <code>#intersex</code></li>
        </ul>
    </div>

    <!-- Быстрый старт -->
    <h2 id="quickstart"><span class="emoji">🚀</span> Быстрый старт</h2>
    <p>Установите интерпретатор и запустите первый скрипт:</p>
    <pre><span class="comment"># hello.rainbow</span>
<span class="keyword">lesbian</span> name = <span class="string">"Мир"</span>;
<span class="keyword">comingout</span> <span class="string">"Привет, "</span> + name + <span class="string">"!"</span>;</pre>
    <p>Запуск:</p>
    <pre>rb10.exe hello.rainbow</pre>
    <p>Или выполните код напрямую:</p>
    <pre>rb10.exe -c 'comingout "🌈 Hello from LGBTScript!";'</pre>

    <!-- Типы данных -->
    <h2 id="types"><span class="emoji">🏷️</span> Типы данных</h2>
    <p>LGBTScript использует инклюзивные ключевые слова для обозначения типов:</p>

    <div class="grid-2">
        <div class="card">
            <strong><code>lesbian</code></strong> — строка<br>
            <span style="color: #6c5b7b; font-size: 0.9rem;">Текстовые данные</span>
        </div>
        <div class="card">
            <strong><code>gay</code></strong> — целое число<br>
            <span style="color: #6c5b7b; font-size: 0.9rem;">Целочисленные значения</span>
        </div>
        <div class="card">
            <strong><code>trans</code></strong> — число с плавающей точкой<br>
            <span style="color: #6c5b7b; font-size: 0.9rem;">Дробные числа</span>
        </div>
        <div class="card">
            <strong><code>nonbinary</code></strong> — булево значение<br>
            <span style="color: #6c5b7b; font-size: 0.9rem;"><code>true</code> или <code>false</code></span>
        </div>
    </div>

    <h3>Литералы</h3>
    <ul>
        <li><strong>Строки:</strong> <code>"Привет, мир!"</code></li>
        <li><strong>Числа:</strong> <code>42</code>, <code>3.14</code></li>
        <li><strong>Булевы:</strong> <code>true</code>, <code>false</code></li>
        <li><strong>Массивы:</strong> <code>[1, 2, 3]</code>, <code>["a", "b", "c"]</code></li>
    </ul>

    <!-- Переменные -->
    <h2 id="variables"><span class="emoji">📦</span> Переменные</h2>
    <p>Переменные объявляются с указанием типа. Значение по умолчанию зависит от типа:</p>
    <ul>
        <li><code>lesbian</code> → <code>""</code> (пустая строка)</li>
        <li><code>gay</code> → <code>0</code></li>
        <li><code>trans</code> → <code>0.0</code></li>
        <li><code>nonbinary</code> → <code>false</code></li>
    </ul>

    <h3>Примеры</h3>
    <pre><span class="keyword">lesbian</span> name = <span class="string">"Alice"</span>;
<span class="keyword">gay</span> age = <span class="number">30</span>;
<span class="keyword">trans</span> pi = <span class="number">3.14159</span>;
<span class="keyword">nonbinary</span> isReady = <span class="keyword">true</span>;

<span class="comment"># Объявление без инициализации (получает значение по умолчанию)</span>
<span class="keyword">lesbian</span> message;
<span class="keyword">comingout</span> message;  <span class="comment"># выведет пустую строку</span>

<span class="comment"># Присваивание (тип должен совпадать)</span>
name = <span class="string">"Bob"</span>;
age = age + <span class="number">1</span>;</pre>

    <!-- Операторы -->
    <h2 id="operators"><span class="emoji">➕</span> Операторы</h2>
    <h3>Арифметические</h3>
    <p><code>+</code> <code>-</code> <code>*</code> <code>/</code> <code>%</code> (для целых чисел)</p>

    <h3>Сравнения</h3>
    <p><code>==</code> <code>!=</code> <code>&lt;</code> <code>&gt;</code> <code>&lt;=</code> <code>&gt;=</code></p>

    <h3>Логические</h3>
    <p><code>&amp;&amp;</code> <code>||</code></p>

    <h3>Конкатенация строк</h3>
    <p>Оператор <code>+</code> работает и для строк:</p>
    <pre><span class="keyword">lesbian</span> greeting = <span class="string">"Hello, "</span> + <span class="string">"World!"</span>;</pre>

    <!-- Массивы -->
    <h2 id="arrays"><span class="emoji">📊</span> Массивы</h2>
    <p>Массивы могут содержать любые значения, включая смешанные типы.</p>
    <pre><span class="keyword">gay</span> numbers = [<span class="number">1</span>, <span class="number">2</span>, <span class="number">3</span>, <span class="number">4</span>, <span class="number">5</span>];
<span class="keyword">lesbian</span> fruits = [<span class="string">"apple"</span>, <span class="string">"banana"</span>, <span class="string">"cherry"</span>];

<span class="comment"># Доступ по индексу</span>
<span class="keyword">comingout</span> numbers[<span class="number">0</span>];  <span class="comment"># 1</span>

<span class="comment"># Присваивание по индексу</span>
numbers[<span class="number">2</span>] = <span class="number">99</span>;

<span class="comment"># Встроенные функции для работы с массивами</span>
append(numbers, <span class="number">6</span>);        <span class="comment"># добавляет элемент</span>
<span class="keyword">gay</span> len = length(numbers);  <span class="comment"># длина массива</span>
remove(numbers, <span class="number">0</span>);           <span class="comment"># удаляет элемент по индексу</span></pre>

    <!-- Управление потоком -->
    <h2 id="control"><span class="emoji">🔄</span> Управление потоком</h2>

    <h3>Условный оператор <code>gender</code></h3>
    <pre><span class="keyword">gender</span> (age >= <span class="number">18</span>) {
    <span class="keyword">comingout</span> <span class="string">"Вы совершеннолетний"</span>;
} <span class="keyword">queer</span> {
    <span class="keyword">comingout</span> <span class="string">"Вы несовершеннолетний"</span>;
}</pre>

    <h3>Цикл <code>pride</code></h3>
    <pre><span class="keyword">gay</span> i = <span class="number">0</span>;
<span class="keyword">pride</span> (i < <span class="number">5</span>) {
    <span class="keyword">comingout</span> <span class="string">"Счёт: "</span> + i;
    i = i + <span class="number">1</span>;
}</pre>

    <!-- Функции -->
    <h2 id="functions"><span class="emoji">🧩</span> Функции</h2>
    <p>Функции объявляются с ключевым словом <code>rainbow</code>. Экспорт — через <code>export</code>.</p>

    <pre><span class="keyword">rainbow</span> greet(name) {
    <span class="keyword">return</span> <span class="string">"Hello, "</span> + name + <span class="string">"!"</span>;
}

<span class="keyword">export</span> <span class="keyword">rainbow</span> calculate(a, b) {
    <span class="keyword">return</span> a + b;
}

<span class="comment"># Вызов</span>
<span class="keyword">lesbian</span> msg = greet(<span class="string">"Alice"</span>);
<span class="keyword">comingout</span> msg;</pre>

    <!-- Встроенные функции -->
    <h2 id="builtins"><span class="emoji">🧰</span> Встроенные функции</h2>

    <h3>Файловая система</h3>
    <ul>
        <li><code>readFile(filename)</code> — читает файл, возвращает строку</li>
        <li><code>writeFile(filename, content)</code> — записывает данные в файл</li>
        <li><code>fileExists(filename)</code> — проверяет существование файла</li>
        <li><code>getDirFiles(path)</code> — возвращает список файлов в директории</li>
    </ul>

    <h3>Работа со строками</h3>
    <ul>
        <li><code>split(text, delimiter)</code> — разбивает строку на массив</li>
        <li><code>replace(text, old, new)</code> — заменяет подстроки</li>
        <li><code>trim(text)</code> — удаляет пробелы в начале и конце</li>
        <li><code>length(value)</code> — длина строки или массива</li>
        <li><code>toUpper(text)</code> / <code>toLower(text)</code> — регистр</li>
    </ul>

    <h3>Регулярные выражения</h3>
    <ul>
        <li><code>regexFind(pattern, text)</code> — находит все совпадения</li>
        <li><code>regexReplace(pattern, text, replacement)</code> — заменяет по шаблону</li>
    </ul>

    <h3>HTTP</h3>
    <ul>
        <li><code>httpGet(url)</code> — GET-запрос</li>
        <li><code>httpPost(url, data)</code> — POST-запрос (JSON)</li>
    </ul>

    <h3>JSON</h3>
    <ul>
        <li><code>jsonParse(jsonString)</code> — парсит JSON в объект/массив</li>
    </ul>

    <h3>Криптография</h3>
    <ul>
        <li><code>md5(text)</code> — MD5-хеш</li>
        <li><code>sha256(text)</code> — SHA256-хеш</li>
    </ul>

    <h3>Система</h3>
    <ul>
        <li><code>getTime()</code> — текущая дата и время</li>
        <li><code>getYear()</code> — текущий год</li>
        <li><code>getMonth()</code> — текущий месяц (1–12)</li>
        <li><code>getOS()</code> — имя операционной системы</li>
        <li><code>getHostname()</code> — имя хоста</li>
        <li><code>getArgs()</code> — аргументы командной строки</li>
        <li><code>hasFlag(flag)</code> — проверяет наличие флага</li>
    </ul>

    <h3>Математика</h3>
    <ul>
        <li><code>random(min, max)</code> — случайное целое число</li>
        <li><code>max(...)</code> / <code>min(...)</code> — максимум/минимум</li>
        <li><code>sqrt(number)</code> — квадратный корень</li>
        <li><code>pow(base, exp)</code> — возведение в степень</li>
    </ul>

    <h3>Прочее</h3>
    <ul>
        <li><code>sleep(ms)</code> — пауза в миллисекундах</li>
        <li><code>append(array, element)</code> — добавляет элемент в массив</li>
        <li><code>remove(array, index)</code> — удаляет элемент по индексу</li>
        <li><code>sendEmail(to, subject, body)</code> — эмуляция отправки письма</li>
    </ul>

    <h3>Пример использования</h3>
    <pre><span class="keyword">lesbian</span> data = readFile(<span class="string">"config.json"</span>);
<span class="keyword">lesbian</span> config = jsonParse(data);
<span class="keyword">comingout</span> <span class="string">"URL: "</span> + config[<span class="string">"api_url"</span>];

<span class="keyword">lesbian</span> response = httpGet(config[<span class="string">"api_url"</span>]);
writeFile(<span class="string">"response.txt"</span>, response);

<span class="keyword">lesbian</span> hash = md5(response);
<span class="keyword">comingout</span> <span class="string">"MD5: "</span> + hash;</pre>

    <!-- Обработка ошибок -->
    <h2 id="error"><span class="emoji">🛡️</span> Обработка ошибок</h2>
    <p>Конструкция <code>try</code> / <code>catch</code> перехватывает ошибки. В блоке <code>catch</code> доступна переменная <code>error</code>.</p>
    <pre><span class="keyword">try</span> {
    <span class="keyword">lesbian</span> result = httpGet(<span class="string">"https://invalid.url"</span>);
    <span class="keyword">comingout</span> result;
} <span class="keyword">catch</span> {
    <span class="keyword">comingout</span> <span class="string">"Ошибка: "</span> + error;
}</pre>

    <!-- Библиотеки -->
    <h2 id="libraries"><span class="emoji">📚</span> Библиотеки</h2>
    <p>Подключение внешних библиотек через директиву <code>#intersex</code>:</p>
    <pre><span class="comment"># Подключение библиотеки</span>
<span class="keyword">#intersex</span> <span class="string">"math.rainbow"</span>
<span class="keyword">#intersex</span> <span class="string">"io.rainbow"</span></pre>
    <p>Поиск библиотек происходит в текущей директории, директории исполняемого файла и папках <code>libs</code>, <code>libraries</code>.</p>

    <!-- CLI -->
    <h2 id="cli"><span class="emoji">💻</span> Интерфейс командной строки</h2>
    <pre>rb10.exe [опции] [файл]

Опции:
  -c "код"          Выполнить код из командной строки
  -lgbt файл        Исполнить файл (аналог позиционного аргумента)
  -tokens           Показать токены после лексического анализа
  -ast              Показать AST после парсинга
  -debug            Включить режим отладки
  -example          Показать расширенный пример

Примеры:
  rb10.exe script.rainbow
  rb10.exe -c 'comingout "Hello!";'
  rb10.exe -lgbt script.rainbow -tokens</pre>

    <!-- Примеры -->
    <h2 id="examples"><span class="emoji">📝</span> Полные примеры</h2>

    <h3>Простой пример</h3>
    <pre><span class="keyword">lesbian</span> name = <span class="string">"World"</span>;
<span class="keyword">comingout</span> <span class="string">"Hello, "</span> + name + <span class="string">"!"</span>;

<span class="keyword">gay</span> x = <span class="number">10</span>;
<span class="keyword">gay</span> y = <span class="number">20</span>;
<span class="keyword">comingout</span> <span class="string">"Sum: "</span> + (x + y);</pre>

    <h3>Работа с файлами и HTTP</h3>
    <pre><span class="keyword">#intersex</span> <span class="string">"io.rainbow"</span>

<span class="keyword">lesbian</span> url = <span class="string">"https://api.github.com"</span>;
<span class="keyword">try</span> {
    <span class="keyword">lesbian</span> response = httpGet(url);
    writeFile(<span class="string">"github.json"</span>, response);
    <span class="keyword">comingout</span> <span class="string">"Данные сохранены"</span>;
} <span class="keyword">catch</span> {
    <span class="keyword">comingout</span> <span class="string">"Ошибка: "</span> + error;
}</pre>

    <h3>Функции и массивы</h3>
    <pre><span class="keyword">export</span> <span class="keyword">rainbow</span> processArray(arr) {
    <span class="keyword">gay</span> sum = <span class="number">0</span>;
    <span class="keyword">gay</span> i = <span class="number">0</span>;
    <span class="keyword">pride</span> (i < length(arr)) {
        sum = sum + arr[i];
        i = i + <span class="number">1</span>;
    }
    <span class="keyword">return</span> sum;
}

<span class="keyword">gay</span> numbers = [<span class="number">1</span>, <span class="number">2</span>, <span class="number">3</span>, <span class="number">4</span>, <span class="number">5</span>];
<span class="keyword">gay</span> total = processArray(numbers);
<span class="keyword">comingout</span> <span class="string">"Сумма: "</span> + total;</pre>

    <hr>

    <div class="footer">
        🌈 LGBTScript — язык, где каждый может найти себя<br>
        <span class="rainbow">Сделано с любовью для всех</span>
    </div>

</div>
</body>
</html>
