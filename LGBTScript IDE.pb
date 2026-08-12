;===============================================================================
; LGBTScript IDE v9.0 "Pride Edition" 
EnableExplicit

CompilerIf #PB_Compiler_OS = #PB_OS_Windows
  Global FontName$ = "Consolas"
  Global UIFont$   = "Segoe UI"
CompilerElse
  Global FontName$ = "Monospace"
  Global UIFont$   = "Sans"
CompilerEndIf

;===============================================================================
; ПАЛИТРА VS CODE DARK+ (с ЛГБТ-акцентами)
;===============================================================================
; Основные цвета VS Code Dark+
#VS_BG            = $1E1E1E        ; Основной фон
#VS_EDITOR_BG     = $1E1E1E        ; Фон редактора
#VS_SIDEBAR_BG    = $252526        ; Боковая панель
#VS_ACTIVITY_BG   = $333333        ; Панель активности
#VS_STATUS_BG     = $007ACC        ; Статус бар (синий)
#VS_INPUT_BG      = $3C3C3C        ; Поля ввода
#VS_TEXT          = $D4D4D4        ; Основной текст
#VS_TEXT_DIM      = $858585        ; Приглушенный текст
#VS_SELECTION     = $264F78        ; Выделение
#VS_CURRENT_LINE  = $2A2D2E        ; Текущая строка
#VS_CARET         = $FFFFFF        ; Каретка

; ЛГБТ-акценты (приглушенные, но заметные)
#PR_RED    = $CC3333        ; Красный
#PR_ORANGE = $CC7733        ; Оранжевый
#PR_YELLOW = $CCCC33        ; Желтый
#PR_GREEN  = $33CC33        ; Зеленый
#PR_BLUE   = $3377CC        ; Синий
#PR_VIOLET = $8833CC        ; Фиолетовый
#PR_PINK   = $CC33AA        ; Розовый
#PR_CYAN   = $33CCCC        ; Бирюзовый

; Стили токенов как в VS Code
#STYLE_KEYWORD    = 1
#STYLE_FUNCTION   = 2
#STYLE_STRING     = 3
#STYLE_COMMENT    = 4
#STYLE_NUMBER     = 5
#STYLE_OPERATOR   = 6
#STYLE_VARIABLE   = 7
#STYLE_PARAMETER  = 8
#STYLE_TYPE       = 9
#STYLE_CONSTANT   = 10
#STYLE_BUILTIN    = 11
#STYLE_KEYWORD_GAY = 12      ; Добавлено для gay
#STYLE_KEYWORD_LESBIAN = 13   ; Добавлено для lesbian, cis, nocis, trans
#STYLE_KEYWORD_GENDER = 14    ; Добавлено для gender
#STYLE_KEYWORD_PRIDE=15
;--- Индикаторы / маркеры ----------------------------------------------------
#IND_WORD     = 8
#IND_ERROR    = 9
#MARKER_ERROR = 1

;--- Уровни лога -------------------------------------------------------------
#LOG_INFO  = 0
#LOG_OK    = 1
#LOG_WARN  = 2
#LOG_ERROR = 3

;--- Гаджеты -----------------------------------------------------------------
#GAD_EDITOR  = 0
#GAD_LOG     = 321
#GAD_SPLIT   = 500
#GAD_TOOLBAR = 501
#GAD_STATUS  = 502

;--- Команды -----------------------------------------------------------------
#CMD_OPEN    = 1
#CMD_SAVE    = 2
#CMD_SAVEAS  = 3
#CMD_EXIT    = 4
#CMD_CUT     = 8
#CMD_COPY    = 9
#CMD_PASTE   = 10
#CMD_ABOUT   = 11
#CMD_UNDO    = 12
#CMD_REDO    = 13
#CMD_RUN     = 20
#CMD_CLEAR   = 21
#CMD_COMPILE = 22       ; Новая команда для компиляции

;--- Геометрия ---------------------------------------------------------------
#TOOLBAR_H    = 48
#STATUS_H     = 28
#STRIPE_H     = 3               ; Тонкая полоса
#EDITOR_RATIO = 70

;--- Совместимость констант Scintilla ----------------------------------------
CompilerIf Not Defined(INDIC_PLAIN, #PB_Constant)
  #INDIC_PLAIN = 0
CompilerEndIf
CompilerIf Not Defined(INDIC_SQUIGGLE, #PB_Constant)
  #INDIC_SQUIGGLE = 1
CompilerEndIf
CompilerIf Not Defined(SCI_SETINDICATORCURRENT, #PB_Constant)
  #SCI_SETINDICATORCURRENT = 2500
CompilerEndIf
CompilerIf Not Defined(SCI_INDICATORFILLRANGE, #PB_Constant)
  #SCI_INDICATORFILLRANGE = 2504
CompilerEndIf
CompilerIf Not Defined(SCI_INDICATORCLEARRANGE, #PB_Constant)
  #SCI_INDICATORCLEARRANGE = 2505
CompilerEndIf
CompilerIf Not Defined(SCI_INDICSETALPHA, #PB_Constant)
  #SCI_INDICSETALPHA = 2523
CompilerEndIf
CompilerIf Not Defined(SCI_INDICSETUNDER, #PB_Constant)
  #SCI_INDICSETUNDER = 2510
CompilerEndIf
CompilerIf Not Defined(SCI_ENSUREVISIBLEENFORCEPOLICY, #PB_Constant)
  #SCI_ENSUREVISIBLEENFORCEPOLICY = 2234
CompilerEndIf
CompilerIf Not Defined(SCI_MARKERDEFINE, #PB_Constant)
  #SCI_MARKERDEFINE = 2040
CompilerEndIf
CompilerIf Not Defined(SCI_MARKERSETFORE, #PB_Constant)
  #SCI_MARKERSETFORE = 2041
CompilerEndIf
CompilerIf Not Defined(SCI_MARKERSETBACK, #PB_Constant)
  #SCI_MARKERSETBACK = 2042
CompilerEndIf
CompilerIf Not Defined(SCI_MARKERADD, #PB_Constant)
  #SCI_MARKERADD = 2043
CompilerEndIf
CompilerIf Not Defined(SCI_MARKERDELETEALL, #PB_Constant)
  #SCI_MARKERDELETEALL = 2045
CompilerEndIf
CompilerIf Not Defined(SC_MARK_ROUNDRECT, #PB_Constant)
  #SC_MARK_ROUNDRECT = 1
CompilerEndIf
CompilerIf Not Defined(SCI_SETMARGINMASKN, #PB_Constant)
  #SCI_SETMARGINMASKN = 2244
CompilerEndIf
CompilerIf Not Defined(SCI_SETCARETLINEVISIBLE, #PB_Constant)
  #SCI_SETCARETLINEVISIBLE = 2096
CompilerEndIf
CompilerIf Not Defined(SCI_SETCARETLINEBACK, #PB_Constant)
  #SCI_SETCARETLINEBACK = 2098
CompilerEndIf
CompilerIf Not Defined(SCI_SETSELBACK, #PB_Constant)
  #SCI_SETSELBACK = 2068
CompilerEndIf
CompilerIf Not Defined(SCI_SETCARETWIDTH, #PB_Constant)
  #SCI_SETCARETWIDTH = 2188
CompilerEndIf
CompilerIf Not Defined(SCI_SETEXTRAASCENT, #PB_Constant)
  #SCI_SETEXTRAASCENT = 2525
CompilerEndIf
CompilerIf Not Defined(SCI_SETEXTRADESCENT, #PB_Constant)
  #SCI_SETEXTRADESCENT = 2527
CompilerEndIf
CompilerIf Not Defined(SCI_SETTABWIDTH, #PB_Constant)
  #SCI_SETTABWIDTH = 2036
CompilerEndIf
CompilerIf Not Defined(SCI_SETUSETABS, #PB_Constant)
  #SCI_SETUSETABS = 2124
CompilerEndIf
CompilerIf Not Defined(SCI_SETINDENTATIONGUIDES, #PB_Constant)
  #SCI_SETINDENTATIONGUIDES = 2132
CompilerEndIf
CompilerIf Not Defined(SC_IV_LOOKBOTH, #PB_Constant)
  #SC_IV_LOOKBOTH = 3
CompilerEndIf
CompilerIf Not Defined(STYLE_INDENTGUIDE, #PB_Constant)
  #STYLE_INDENTGUIDE = 37
CompilerEndIf
CompilerIf Not Defined(SCI_SETEOLMODE, #PB_Constant)
  #SCI_SETEOLMODE = 2031
CompilerEndIf
CompilerIf Not Defined(SC_EOL_LF, #PB_Constant)
  #SC_EOL_LF = 2
CompilerEndIf
CompilerIf Not Defined(SCI_SETSCROLLWIDTHTRACKING, #PB_Constant)
  #SCI_SETSCROLLWIDTHTRACKING = 2516
CompilerEndIf
CompilerIf Not Defined(SCI_SETSCROLLWIDTH, #PB_Constant)
  #SCI_SETSCROLLWIDTH = 2514
CompilerEndIf
CompilerIf Not Defined(SCI_SETWRAPMODE, #PB_Constant)
  #SCI_SETWRAPMODE = 2268
CompilerEndIf
CompilerIf Not Defined(SCI_SETHSCROLLBAR, #PB_Constant)
  #SCI_SETHSCROLLBAR = 2180
CompilerEndIf
CompilerIf Not Defined(SC_WRAP_WORD, #PB_Constant)
  #SC_WRAP_WORD = 1
CompilerEndIf

;===============================================================================
; Структуры
;===============================================================================
Structure TBtn
  x.i
  y.i
  w.i
  h.i
  cmd.i
  text$
  accent.i
  hover.i
  down.i
  disabled.i
EndStructure

;--- Глобальные данные -------------------------------------------------------
Global NewMap Keywords.i()
Global NewMap Builtins.i()
Global NewMap DeclaredVariables.i()
Global NewMap DeclaredFunctions.i()
Global NewMap LogLineTarget.i()
Global NewList TB.TBtn()

Global *TextBuf
Global TextBufLen.i = 0

Global IsHighlighting.i = #False
Global Quit.i = 0
Global CurrentFile$ = ""

Global WordIndStart.i = -1
Global WordIndEnd.i   = -1

Global LastExitCode.i  = 0
Global LastErrCount.i  = 0
Global LastWarnCount.i = 0

Global AsyncInProgress.i  = #False
Global AsyncCurrentLine.i = 0
Global AsyncTotalLines.i  = 0
Global AsyncGadget.i      = 0
Global AsyncChunkLines.i  = 400

Global FontUI.i, FontUIBold.i, FontMono.i, FontTitle.i
Global StatusLeft$  = "Готов к работе"
Global StatusRight$ = ""
Global StatusMode.i = 0

#TIMER_HIGHLIGHT = 100
#TIMER_ASYNC     = 101
#ASYNC_THRESHOLD = 1048576

;--- Декларации --------------------------------------------------------------
Declare   HighlightText(Gadget)
Declare   StartAsyncHighlight(Gadget)
Declare   AsyncHighlightStep()
Declare   StyleRange(Gadget, startPos, endPos)
Declare   CollectDeclarations()
Declare   UpdateTextBuffer(Gadget)
Declare   UpdateWordUnderCaret(Gadget)
Declare.i StopAsync()
Declare.i LogLine(text$, kind = #LOG_INFO)
Declare   LogError(text$)
Declare.i ParseErrorLine(text$)
Declare   GotoEditorLine(lineNo)
Declare   MarkErrorLine(lineNo)
Declare   ClearErrorMarks()
Declare   RunRainbow()
Declare   CompileToExe()      ; Новая функция для компиляции
Declare.i SaveEditorTo(File$)
Declare   LoadFileToEditor(File$)
Declare   LayoutUI()
Declare   Status(text$, mode = 0)
Declare   UpdateCaretStatus()
Declare   DrawToolbar()
Declare   DrawStatusBar()
Declare   DoCommand(cmd)

;===============================================================================
; РАДУЖНАЯ ГРАФИКА
;===============================================================================
Procedure.i MixColor(c1, c2, f.f)
  Protected r, g, b
  r = Red(c1)   + (Red(c2)   - Red(c1))   * f
  g = Green(c1) + (Green(c2) - Green(c1)) * f
  b = Blue(c1)  + (Blue(c2)  - Blue(c1))  * f
  ProcedureReturn RGB(r, g, b)
EndProcedure

Procedure.i RainbowAt(t.f)
  Protected Dim st(5)
  Protected seg.f, idx, f.f

  st(0) = #PR_RED
  st(1) = #PR_ORANGE
  st(2) = #PR_YELLOW
  st(3) = #PR_GREEN
  st(4) = #PR_BLUE
  st(5) = #PR_VIOLET

  If t < 0.0 : t = 0.0 : EndIf
  If t > 1.0 : t = 1.0 : EndIf

  seg = t * 5.0
  idx = Int(seg)
  If idx > 4 : idx = 4 : EndIf
  f = seg - idx

  ProcedureReturn MixColor(st(idx), st(idx + 1), f)
EndProcedure

Procedure DrawRainbowH(x, y, w, h, k.f = 1.0, bg = #VS_BG)
  Protected i, c
  If w <= 0 Or h <= 0 : ProcedureReturn : EndIf
  For i = 0 To w - 1
    c = RainbowAt(i / (w - 1.0))
    If k < 1.0 : c = MixColor(bg, c, k) : EndIf
    Line(x + i, y, 1, h, c)
  Next
EndProcedure

;===============================================================================
; Тёмный заголовок окна
;===============================================================================
CompilerIf #PB_Compiler_OS = #PB_OS_Windows
Procedure StylizeTitleBar(win)
  Protected dark.l = 1, lib
  Protected cap.l  = #VS_BG
  Protected txt.l  = #VS_TEXT
  Protected brd.l  = #PR_VIOLET

  lib = OpenLibrary(#PB_Any, "dwmapi.dll")
  If lib
    CallFunction(lib, "DwmSetWindowAttribute", WindowID(win), 20, @dark, 4)
    CallFunction(lib, "DwmSetWindowAttribute", WindowID(win), 19, @dark, 4)
    CallFunction(lib, "DwmSetWindowAttribute", WindowID(win), 35, @cap,  4)
    CallFunction(lib, "DwmSetWindowAttribute", WindowID(win), 36, @txt,  4)
    CallFunction(lib, "DwmSetWindowAttribute", WindowID(win), 34, @brd,  4)
    CloseLibrary(lib)
  EndIf
EndProcedure
CompilerEndIf

;===============================================================================
; Инициализация словарей
;===============================================================================
Procedure InitKeywords()
  ClearMap(Keywords())
  Keywords("gay")        = #STYLE_KEYWORD
  Keywords("homo")        = #STYLE_KEYWORD
    Keywords("pride")        = #STYLE_KEYWORD
   Keywords("queer")        = #STYLE_KEYWORD
  Keywords("lesbian")    = #STYLE_KEYWORD
  Keywords("trans")      = #STYLE_KEYWORD
  Keywords("nocis")      = #STYLE_KEYWORD
  Keywords("pride")      = #STYLE_KEYWORD
  Keywords("rainbow")    = #STYLE_KEYWORD
  Keywords("cis")        = #STYLE_KEYWORD
  Keywords("comingout")  = #STYLE_KEYWORD
  Keywords("nonbinary")  = #STYLE_KEYWORD
  Keywords("help")       = #STYLE_KEYWORD
  Keywords("orientation")= #STYLE_KEYWORD
  Keywords("gender")     = #STYLE_KEYWORD
  Keywords("return")     = #STYLE_KEYWORD
  Keywords("true")       = #STYLE_KEYWORD
  Keywords("false")      = #STYLE_KEYWORD
  Keywords("try")        = #STYLE_KEYWORD
  Keywords("catch")      = #STYLE_KEYWORD
  Keywords("export")     = #STYLE_KEYWORD
EndProcedure

Procedure InitBuiltinFunctions()
  Protected List$, i, w$
  ClearMap(Builtins())
  List$ = "readFile writeFile fileExists getDirFiles " +
          "split replace trim length toUpper toLower " +
          "append remove " +
          "random max min sqrt pow " +
          "getTime getYear getMonth " +
          "getOS getHostname getArgs hasFlag " +
          "httpGet httpPost jsonParse md5 sha256 " +
          "regexFind regexReplace sleep sendEmail"
  For i = 1 To CountString(List$, " ") + 1
    w$ = StringField(List$, i, " ")
    If w$ <> "" : Builtins(w$) = 1 : EndIf
  Next
EndProcedure

;===============================================================================
; Проверки символов
;===============================================================================
Procedure.i IsIdentStart(c.a)
  If (c >= 'A' And c <= 'Z') Or (c >= 'a' And c <= 'z') Or c = '_' Or c >= 128
    ProcedureReturn #True
  EndIf
  ProcedureReturn #False
EndProcedure

Procedure.i IsIdentChar(c.a)
  If IsIdentStart(c) Or (c >= '0' And c <= '9') : ProcedureReturn #True : EndIf
  ProcedureReturn #False
EndProcedure

Procedure.i IsOperatorChar(c.a)
  Select c
    Case '+', '-', '*', '/', '=', '<', '>', '&', '|', '!', '^', '%', '~', '?',
         ':', ';', ',', '.', '[', ']', '{', '}'
      ProcedureReturn #True
  EndSelect
  ProcedureReturn #False
EndProcedure

;===============================================================================
; Буфер документа
;===============================================================================
Procedure UpdateTextBuffer(Gadget)
  If *TextBuf
    FreeMemory(*TextBuf)
    *TextBuf = 0
  EndIf
  TextBufLen = ScintillaSendMessage(Gadget, #SCI_GETLENGTH, 0, 0)
  If TextBufLen <= 0
    TextBufLen = 0
    ProcedureReturn
  EndIf
  *TextBuf = AllocateMemory(TextBufLen + 1)
  If *TextBuf
    ScintillaSendMessage(Gadget, #SCI_GETTEXT, TextBufLen + 1, *TextBuf)
  Else
    TextBufLen = 0
  EndIf
EndProcedure

Procedure.s ReadWord2(startPos, lengthBytes)
  If lengthBytes <= 0 Or Not *TextBuf : ProcedureReturn "" : EndIf
  ProcedureReturn PeekS(*TextBuf + startPos, lengthBytes, #PB_UTF8 | #PB_ByteLength)
EndProcedure

;===============================================================================
; Сбор объявлений
;===============================================================================
Procedure CollectDeclarations()
  Protected pos, c.a, wStart, w$, pending

  ClearMap(DeclaredVariables())
  ClearMap(DeclaredFunctions())
  If Not *TextBuf Or TextBufLen = 0 : ProcedureReturn : EndIf

  pos = 0
  pending = 0
  While pos < TextBufLen
    c = PeekA(*TextBuf + pos)

    If c = '@'
      While pos < TextBufLen
        c = PeekA(*TextBuf + pos)
        If c = 10 Or c = 13 : Break : EndIf
        pos + 1
      Wend
      pending = 0
      Continue
    EndIf

    If c = '"'
      pos + 1
      While pos < TextBufLen
        c = PeekA(*TextBuf + pos)
        If c = 10 Or c = 13 : Break : EndIf
        If c = '\' And pos + 1 < TextBufLen
          pos + 2
          Continue
        EndIf
        pos + 1
        If c = '"' : Break : EndIf
      Wend
      pending = 0
      Continue
    EndIf

    If IsIdentStart(c)
      wStart = pos
      While pos < TextBufLen And IsIdentChar(PeekA(*TextBuf + pos))
        pos + 1
      Wend
      w$ = ReadWord2(wStart, pos - wStart)

      Select LCase(w$)
        Case "gay", "lesbian", "trans", "nonbinary", "gender", "queer", "homo", "asexual"
          pending = 2
        Case "rainbow"
          pending = 2
        Default
          If pending = 1 And Not FindMapElement(Keywords(), LCase(w$))
            DeclaredVariables(w$) = 1
          ElseIf pending = 2 And Not FindMapElement(Keywords(), LCase(w$))
            DeclaredFunctions(w$) = 1
          EndIf
          pending = 0
      EndSelect
      Continue
    EndIf

    If c <> ' ' And c <> 9 : pending = 0 : EndIf
    pos + 1
  Wend
EndProcedure

;===============================================================================
; Стиль слова (МОДИФИЦИРОВАНА)
;===============================================================================
Procedure.i WordStyle(w$, nextIsParen)
  Protected lw$ = LCase(w$)

  ; Проверяем конкретные ключевые слова с особыми стилями
  Select lw$
    Case "gay"
      ProcedureReturn #STYLE_KEYWORD_GAY
    Case "lesbian", "cis", "nocis", "trans", "sex"
      ProcedureReturn #STYLE_KEYWORD_LESBIAN
    Case "gender", "asexual"
      ProcedureReturn #STYLE_KEYWORD_GENDER
       Case "homo"
      ProcedureReturn #STYLE_KEYWORD_GAY
   
  EndSelect

  ; Остальные ключевые слова используют общий стиль
  If FindMapElement(Keywords(), lw$) : ProcedureReturn #STYLE_KEYWORD : EndIf
  If FindMapElement(Builtins(), w$) : ProcedureReturn #STYLE_BUILTIN : EndIf
  If FindMapElement(DeclaredFunctions(), w$) : ProcedureReturn #STYLE_FUNCTION : EndIf
  If FindMapElement(DeclaredVariables(), w$) : ProcedureReturn #STYLE_VARIABLE : EndIf
  If nextIsParen : ProcedureReturn #STYLE_FUNCTION : EndIf
  ProcedureReturn #VS_TEXT
EndProcedure

;===============================================================================
; Подчёркивание слова под кареткой
;===============================================================================
Procedure UpdateWordUnderCaret(Gadget)
  Protected pos, s, e, w$, st, docLen

  docLen = ScintillaSendMessage(Gadget, #SCI_GETLENGTH, 0, 0)

  If WordIndStart >= 0
    ScintillaSendMessage(Gadget, #SCI_SETINDICATORCURRENT, #IND_WORD, 0)
    ScintillaSendMessage(Gadget, #SCI_INDICATORCLEARRANGE, 0, docLen)
    WordIndStart = -1
    WordIndEnd   = -1
  EndIf

  If Not *TextBuf Or TextBufLen <> docLen Or docLen = 0 : ProcedureReturn : EndIf

  If ScintillaSendMessage(Gadget, #SCI_GETSELECTIONSTART, 0, 0) <>0 
     ScintillaSendMessage(Gadget, #SCI_GETSELECTIONEND, 0, 0)
    ProcedureReturn
  EndIf

  pos = ScintillaSendMessage(Gadget, #SCI_GETCURRENTPOS, 0, 0)
  If pos < 0 Or pos > TextBufLen : ProcedureReturn : EndIf

  s = pos
  While s > 0 And IsIdentChar(PeekA(*TextBuf + s - 1)) : s - 1 : Wend
  e = pos
  While e < TextBufLen And IsIdentChar(PeekA(*TextBuf + e)) : e + 1 : Wend

  If e <= s : ProcedureReturn : EndIf

  st = ScintillaSendMessage(Gadget, #SCI_GETSTYLEAT, s, 0)
  If st = #STYLE_COMMENT Or st = #STYLE_STRING : ProcedureReturn : EndIf

  w$ = ReadWord2(s, e - s)
  If w$ = "" : ProcedureReturn : EndIf

  If FindMapElement(Keywords(), LCase(w$)) Or
     FindMapElement(Builtins(), w$) Or
     FindMapElement(DeclaredFunctions(), w$) Or
     FindMapElement(DeclaredVariables(), w$)

    ScintillaSendMessage(Gadget, #SCI_SETINDICATORCURRENT, #IND_WORD, 0)
    ScintillaSendMessage(Gadget, #SCI_INDICATORFILLRANGE, s, e - s)
    WordIndStart = s
    WordIndEnd   = e
  EndIf
EndProcedure

;===============================================================================
; Лексер
;===============================================================================
Procedure StyleRange(Gadget, startPos, endPos)
  Protected pos, c.a, tokStart, tokLen, w$, st, p, nextIsParen

  If Not *TextBuf Or endPos > TextBufLen : endPos = TextBufLen : EndIf
  If startPos < 0 Or startPos >= endPos : ProcedureReturn : EndIf

  ScintillaSendMessage(Gadget, #SCI_STARTSTYLING, startPos, 0)
  pos = startPos

  While pos < endPos
    c = PeekA(*TextBuf + pos)
    tokStart = pos

    If c = '@'
      While pos < endPos
        c = PeekA(*TextBuf + pos)
        If c = 10 Or c = 13 : Break : EndIf
        pos + 1
      Wend
      ScintillaSendMessage(Gadget, #SCI_SETSTYLING, pos - tokStart, #STYLE_COMMENT)
      Continue
    EndIf

    If c = '"'
      pos + 1
      While pos < endPos
        c = PeekA(*TextBuf + pos)
        If c = 10 Or c = 13 : Break : EndIf
        If c = '\' And pos + 1 < endPos
          pos + 2
          Continue
        EndIf
        pos + 1
        If c = '"' : Break : EndIf
      Wend
      ScintillaSendMessage(Gadget, #SCI_SETSTYLING, pos - tokStart, #STYLE_STRING)
      Continue
    EndIf

    If c = '#'
      pos + 1
      While pos < endPos And IsIdentChar(PeekA(*TextBuf + pos))
        pos + 1
      Wend
      w$ = LCase(ReadWord2(tokStart, pos - tokStart))
      If w$ = "#intersex"
        st = #STYLE_TYPE
      Else
        st = #STYLE_CONSTANT
      EndIf
      ScintillaSendMessage(Gadget, #SCI_SETSTYLING, pos - tokStart, st)
      Continue
    EndIf

    If IsIdentStart(c)
      While pos < endPos And IsIdentChar(PeekA(*TextBuf + pos))
        pos + 1
      Wend
      tokLen = pos - tokStart
      w$ = ReadWord2(tokStart, tokLen)

      nextIsParen = #False
      p = pos
      While p < TextBufLen
        c = PeekA(*TextBuf + p)
        If c = ' ' Or c = 9
          p + 1
          Continue
        EndIf
        If c = '(' : nextIsParen = #True : EndIf
        Break
      Wend

      ScintillaSendMessage(Gadget, #SCI_SETSTYLING, tokLen, WordStyle(w$, nextIsParen))
      Continue
    EndIf

    If c >= '0' And c <= '9'
      While pos < endPos
        c = PeekA(*TextBuf + pos)
        If (c >= '0' And c <= '9') Or c = '.'
          pos + 1
        Else
          Break
        EndIf
      Wend
      ScintillaSendMessage(Gadget, #SCI_SETSTYLING, pos - tokStart, #STYLE_NUMBER)
      Continue
    EndIf

    If c = '(' Or c = ')'
      pos + 1
      ScintillaSendMessage(Gadget, #SCI_SETSTYLING, 1, #STYLE_FUNCTION)
      Continue
    EndIf

    If IsOperatorChar(c)
      pos + 1
      ScintillaSendMessage(Gadget, #SCI_SETSTYLING, 1, #STYLE_OPERATOR)
      Continue
    EndIf

    pos + 1
    ScintillaSendMessage(Gadget, #SCI_SETSTYLING, 1, #VS_TEXT)
  Wend
EndProcedure

;===============================================================================
; Подсветка
;===============================================================================
Procedure HighlightText(Gadget)
  If IsHighlighting : ProcedureReturn : EndIf
  IsHighlighting = #True

  UpdateTextBuffer(Gadget)
  If TextBufLen > 0
    CollectDeclarations()
    StyleRange(Gadget, 0, TextBufLen)
  EndIf

  IsHighlighting = #False
  UpdateWordUnderCaret(Gadget)
EndProcedure

Procedure.i StopAsync()
  If AsyncInProgress
    RemoveWindowTimer(0, #TIMER_ASYNC)
    AsyncInProgress = #False
    ProcedureReturn #True
  EndIf
  ProcedureReturn #False
EndProcedure

Procedure StartAsyncHighlight(Gadget)
  StopAsync()

  UpdateTextBuffer(Gadget)
  If TextBufLen = 0 : ProcedureReturn : EndIf

  CollectDeclarations()

  AsyncTotalLines  = ScintillaSendMessage(Gadget, #SCI_GETLINECOUNT, 0, 0)
  AsyncCurrentLine = 0
  AsyncGadget      = Gadget
  AsyncInProgress  = #True

  AddWindowTimer(0, #TIMER_ASYNC, 10)
  Status("Подсветка: 0%")
EndProcedure

Procedure AsyncHighlightStep()
  Protected endLine, startPos, endPos, percent

  If Not AsyncInProgress
    RemoveWindowTimer(0, #TIMER_ASYNC)
    ProcedureReturn
  EndIf

  endLine = AsyncCurrentLine + AsyncChunkLines - 1
  If endLine >= AsyncTotalLines : endLine = AsyncTotalLines - 1 : EndIf

  startPos = ScintillaSendMessage(AsyncGadget, #SCI_POSITIONFROMLINE, AsyncCurrentLine, 0)
  If endLine + 1 < AsyncTotalLines
    endPos = ScintillaSendMessage(AsyncGadget, #SCI_POSITIONFROMLINE, endLine + 1, 0)
  Else
    endPos = TextBufLen
  EndIf

  StyleRange(AsyncGadget, startPos, endPos)

  AsyncCurrentLine = endLine + 1
  If AsyncCurrentLine >= AsyncTotalLines
    AsyncInProgress = #False
    RemoveWindowTimer(0, #TIMER_ASYNC)
    Status("Готово (" + Str(AsyncTotalLines) + " строк)", 2)
    UpdateWordUnderCaret(AsyncGadget)
  Else
    percent = (AsyncCurrentLine * 100) / AsyncTotalLines
    Status("Подсветка: " + Str(percent) + "%")
  EndIf
EndProcedure

Procedure TimerHandler()
  Select EventTimer()
    Case #TIMER_HIGHLIGHT
      RemoveWindowTimer(0, #TIMER_HIGHLIGHT)
      StopAsync()
      If Not IsHighlighting
        If ScintillaSendMessage(#GAD_EDITOR, #SCI_GETLENGTH, 0, 0) > #ASYNC_THRESHOLD
          StartAsyncHighlight(#GAD_EDITOR)
        Else
          HighlightText(#GAD_EDITOR)
        EndIf
      EndIf

    Case #TIMER_ASYNC
      AsyncHighlightStep()
  EndSelect
EndProcedure

Procedure ScintillaCB(Gadget, *scinotify.SCNotification)
  Select *scinotify\nmhdr\code

    Case #SCN_MODIFIED
      If *scinotify\modificationType & (#SC_MOD_INSERTTEXT | #SC_MOD_DELETETEXT)
        WordIndStart = -1
        WordIndEnd   = -1
        RemoveWindowTimer(0, #TIMER_HIGHLIGHT)
        AddWindowTimer(0, #TIMER_HIGHLIGHT, 150)
      EndIf

    Case #SCN_UPDATEUI
      UpdateWordUnderCaret(Gadget)
      UpdateCaretStatus()

  EndSelect
  ProcedureReturn #True
EndProcedure

;===============================================================================
; Настройка Scintilla — VS Code Dark+ тема (МОДИФИЦИРОВАНА)
;===============================================================================
Procedure SetStyle(Gadget, style, fore, back = #VS_EDITOR_BG, bold = #False, italic = #False)
  ScintillaSendMessage(Gadget, #SCI_STYLESETFORE, style, fore)
  ScintillaSendMessage(Gadget, #SCI_STYLESETBACK, style, back)
  If bold   : ScintillaSendMessage(Gadget, #SCI_STYLESETBOLD,   style, #True) : EndIf
  If italic : ScintillaSendMessage(Gadget, #SCI_STYLESETITALIC, style, #True) : EndIf
EndProcedure

Procedure SetupScintilla(Gadget)
  ScintillaSendMessage(Gadget, #SCI_STYLESETFONT, #STYLE_DEFAULT, @FontName$)
  ScintillaSendMessage(Gadget, #SCI_STYLESETSIZE, #STYLE_DEFAULT, 11)
  ScintillaSendMessage(Gadget, #SCI_STYLESETFORE, #STYLE_DEFAULT, #VS_TEXT)
  ScintillaSendMessage(Gadget, #SCI_STYLESETBACK, #STYLE_DEFAULT, #VS_EDITOR_BG)
  ScintillaSendMessage(Gadget, #SCI_STYLECLEARALL)

  ScintillaSendMessage(Gadget, #SCI_SETCODEPAGE, #SC_CP_UTF8, 0)
  ScintillaSendMessage(Gadget, #SCI_SETEOLMODE, #SC_EOL_LF, 0)

  ;--- Без горизонтальной прокрутки ---
  ScintillaSendMessage(Gadget, #SCI_SETSCROLLWIDTHTRACKING, #False, 0)
  ScintillaSendMessage(Gadget, #SCI_SETSCROLLWIDTH, 1, 0)
  ScintillaSendMessage(Gadget, #SCI_SETWRAPMODE, #SC_WRAP_WORD, 0)
  ScintillaSendMessage(Gadget, #SCI_SETHSCROLLBAR, #False, 0)

  ; поля
  ScintillaSendMessage(Gadget, #SCI_SETMARGINTYPEN,  0, #SC_MARGIN_NUMBER)
  ScintillaSendMessage(Gadget, #SCI_SETMARGINWIDTHN, 0, 54)
  ScintillaSendMessage(Gadget, #SCI_STYLESETBACK, #STYLE_LINENUMBER, #VS_EDITOR_BG)
  ScintillaSendMessage(Gadget, #SCI_STYLESETFORE, #STYLE_LINENUMBER, #VS_TEXT_DIM)

  ScintillaSendMessage(Gadget, #SCI_SETMARGINTYPEN,  1, #SC_MARGIN_SYMBOL)
  ScintillaSendMessage(Gadget, #SCI_SETMARGINWIDTHN, 1, 16)
  ScintillaSendMessage(Gadget, #SCI_SETMARGINMASKN,  1, 1 << #MARKER_ERROR)
  ScintillaSendMessage(Gadget, #SCI_MARKERDEFINE,  #MARKER_ERROR, #SC_MARK_ROUNDRECT)
  ScintillaSendMessage(Gadget, #SCI_MARKERSETFORE, #MARKER_ERROR, #VS_TEXT)
  ScintillaSendMessage(Gadget, #SCI_MARKERSETBACK, #MARKER_ERROR, #PR_RED)

  ; каретка и выделение
  ScintillaSendMessage(Gadget, #SCI_SETCARETFORE, #VS_CARET)
  ScintillaSendMessage(Gadget, #SCI_SETCARETWIDTH, 2, 0)
  ScintillaSendMessage(Gadget, #SCI_SETCARETLINEVISIBLE, #True, 0)
  ScintillaSendMessage(Gadget, #SCI_SETCARETLINEBACK, #VS_CURRENT_LINE, 0)
  ScintillaSendMessage(Gadget, #SCI_SETSELBACK, #True, #VS_SELECTION)

  ; отступы
  ScintillaSendMessage(Gadget, #SCI_SETTABWIDTH, 4, 0)
  ScintillaSendMessage(Gadget, #SCI_SETUSETABS, #False, 0)
  ScintillaSendMessage(Gadget, #SCI_SETINDENTATIONGUIDES, #SC_IV_LOOKBOTH, 0)
  ScintillaSendMessage(Gadget, #SCI_STYLESETFORE, #STYLE_INDENTGUIDE, #VS_TEXT_DIM)
  ScintillaSendMessage(Gadget, #SCI_STYLESETBACK, #STYLE_INDENTGUIDE, #VS_EDITOR_BG)

  ScintillaSendMessage(Gadget, #SCI_SETEXTRAASCENT, 2, 0)
  ScintillaSendMessage(Gadget, #SCI_SETEXTRADESCENT, 2, 0)
  ScintillaSendMessage(Gadget, #SCI_SETLAYOUTCACHE, #SC_CACHE_PAGE, 0)
  ScintillaSendMessage(Gadget, #SCI_SETMODEVENTMASK, #SC_MOD_INSERTTEXT | #SC_MOD_DELETETEXT)

  ; цвета токенов как в VS Code
  SetStyle(Gadget, #STYLE_KEYWORD,            #PR_PINK,           #VS_EDITOR_BG, #True)  ; Ключевые слова - розовые + жирный
  SetStyle(Gadget, #STYLE_FUNCTION,           #PR_YELLOW)         ; Функции - желтые
  SetStyle(Gadget, #STYLE_STRING,             #PR_ORANGE)         ; Строки - оранжевые
  SetStyle(Gadget, #STYLE_COMMENT,            $6A9955, #VS_EDITOR_BG, #False, #True) ; Комментарии - зеленые
  SetStyle(Gadget, #STYLE_NUMBER,             $B5CEA8)            ; Числа - светло-зеленые
  SetStyle(Gadget, #STYLE_OPERATOR,           #VS_TEXT)           ; Операторы - белые
  SetStyle(Gadget, #STYLE_VARIABLE,           #PR_CYAN)           ; Переменные - бирюзовые
  SetStyle(Gadget, #STYLE_PARAMETER,          #PR_BLUE)           ; Параметры - синие
  SetStyle(Gadget, #STYLE_TYPE,               #PR_VIOLET)         ; Типы - фиолетовые
  SetStyle(Gadget, #STYLE_CONSTANT,           #PR_VIOLET)         ; Константы - фиолетовые
  SetStyle(Gadget, #STYLE_BUILTIN,            #PR_GREEN)          ; Встроенные - зеленые

  ; Новые стили для ключевых слов
  SetStyle(Gadget, #STYLE_KEYWORD_GAY,        RGB(0, 190, 255),   #VS_EDITOR_BG, #True)   ; gay - голубой + жирный
  SetStyle(Gadget, #STYLE_KEYWORD_LESBIAN,    RGB(255, 105, 180), #VS_EDITOR_BG, #True)   ; lesbian - розовый + жирный
  SetStyle(Gadget, #STYLE_KEYWORD_GENDER,     RGB(255, 255, 255), #VS_EDITOR_BG, #True)    ; gender - белый + жирный
 SetStyle(Gadget, #STYLE_KEYWORD_PRIDE,     #Yellow, #VS_EDITOR_BG, #True)    ; gender - белый + жирный

  ; индикаторы
  ScintillaSendMessage(Gadget, #SCI_INDICSETSTYLE, #IND_WORD, #INDIC_PLAIN)
  ScintillaSendMessage(Gadget, #SCI_INDICSETFORE,  #IND_WORD, #PR_YELLOW)
  ScintillaSendMessage(Gadget, #SCI_INDICSETALPHA, #IND_WORD, 200)
  ScintillaSendMessage(Gadget, #SCI_INDICSETUNDER, #IND_WORD, #False)

  ScintillaSendMessage(Gadget, #SCI_INDICSETSTYLE, #IND_ERROR, #INDIC_SQUIGGLE)
  ScintillaSendMessage(Gadget, #SCI_INDICSETFORE,  #IND_ERROR, #PR_RED)
  ScintillaSendMessage(Gadget, #SCI_INDICSETUNDER, #IND_ERROR, #False)
EndProcedure

; ТУЛБАР
;===============================================================================
Procedure AddTB(cmd, text$, accent, rainbow = #False)
  AddElement(TB())
  TB()\cmd     = cmd
  TB()\text$   = text$
  TB()\accent  = accent

  TB()\h       = 30
EndProcedure

Procedure InitToolbar()
  ;ClearList(TB())
  
  AddTB(#CMD_OPEN,    "Открыть",      #VS_BG)
  AddTB(#CMD_SAVE,    "Сохранить",    #VS_BG)
  AddTB(#CMD_RUN,     "▶ Выполнить",  #PR_GREEN)
  AddTB(#CMD_COMPILE, "⚙ Компиляция", #PR_BLUE)  ; Новая кнопка для компиляции
  AddTB(#CMD_CLEAR,   "✕ Очистить",   #PR_RED)
  AddTB(#CMD_EXIT,    "Выход",        #PR_RED)
EndProcedure

Procedure DrawToolbar()
  Protected w, h, tw, tx, ty, c, i
  Protected bg, fg

  If Not IsGadget(#GAD_TOOLBAR) : ProcedureReturn : EndIf

  w = GadgetWidth(#GAD_TOOLBAR)
  h = GadgetHeight(#GAD_TOOLBAR)

  If Not StartDrawing(CanvasOutput(#GAD_TOOLBAR)) : ProcedureReturn : EndIf

  ; Благородный фон как в VS Code
  Box(0, 0, w, h, #VS_ACTIVITY_BG)
  DrawRainbowH(0, 0, w, h - #STRIPE_H+30, 0.08, #VS_ACTIVITY_BG)
  DrawRainbowH(0, h - #STRIPE_H+30, w, #STRIPE_H, 0.9)

  If FontUI : DrawingFont(FontID(FontUI)) : EndIf

  ForEach TB()
    bg = MixColor(#VS_ACTIVITY_BG, TB()\accent, 0.30)
    fg = #VS_TEXT

    If TB()
      DrawingMode(#PB_2DDrawing_Default)
      For i = 0 To TB()\w - 1
        c = RainbowAt(i / (TB()\w - 1.0))
        If TB()\disabled
          c = MixColor(#VS_ACTIVITY_BG, c, 0.30)
        ElseIf TB()\down
          c = MixColor(#VS_BG, c, 0.72)
        ElseIf Not TB()\hover
          c = MixColor(#VS_ACTIVITY_BG, c, 0.70)
        EndIf
        Line(TB()\x + i, TB()\y, 1, TB()\h, c)
      Next
      fg = #VS_TEXT
    Else
      If TB()\disabled
        bg = MixColor(#VS_ACTIVITY_BG, TB()\accent, 0.08)
        fg = #VS_TEXT_DIM
      ElseIf TB()\down
        bg = MixColor(#VS_BG, TB()\accent, 0.50)
        
        fg = #VS_TEXT
      ElseIf TB()\hover
        bg = MixColor(#VS_ACTIVITY_BG, TB()\accent, 0.45)
        fg = #VS_TEXT
      EndIf
      DrawingMode(#PB_2DDrawing_Default)
      RoundBox(TB()\x, TB()\y, TB()\w, TB()\h, 4, 4, bg)
    EndIf

    DrawingMode(#PB_2DDrawing_Transparent)
    tw = TextWidth(TB()\text$)
    tx = TB()\x + (TB()\w - tw) / 3
    ty = TB()\y + (TB()\h - TextHeight(TB()\text$)) / 2

    DrawText(tx, ty, TB()\text$, fg)
  Next

  StopDrawing()
EndProcedure

Procedure.i ToolbarHit(mx, my)
  ForEach TB()
    If mx >= TB()\x And mx < TB()\x + TB()\w And
       my >= TB()\y And my < TB()\y + TB()\h
      ProcedureReturn TB()\cmd
    EndIf
  Next
  ProcedureReturn 0
EndProcedure

Procedure ToolbarHover(mx, my)
  Protected changed = #False, nh

  ForEach TB()
    nh = #False
    If mx >= TB()\x And mx < TB()\x + TB()\w And
       my >= TB()\y And my < TB()\y + TB()\h
      nh = #True
    EndIf
    If nh <> TB()\hover
      TB()\hover = nh
      changed = #True
    EndIf
  Next

  If changed : DrawToolbar() : EndIf
EndProcedure

Procedure SetRunEnabled(state)
  ForEach TB()
    If TB()\cmd = #CMD_RUN Or TB()\cmd = #CMD_COMPILE
      TB()\disabled = Bool(Not state)
    EndIf
  Next
  DrawToolbar()
EndProcedure

;===============================================================================
; СТАТУС-СТРОКА
;===============================================================================
Procedure DrawStatusBar()
  Protected w, h, base, tw

  If Not IsGadget(#GAD_STATUS) : ProcedureReturn : EndIf

  w = GadgetWidth(#GAD_STATUS)
  h = GadgetHeight(#GAD_STATUS)

  If Not StartDrawing(CanvasOutput(#GAD_STATUS)) : ProcedureReturn : EndIf

  Select StatusMode
    Case 1 : base = #PR_RED
    Case 2 : base = #PR_GREEN
    Default: base = #VS_STATUS_BG
  EndSelect

  ; Статус бар как в VS Code
  ;Box(0, 0, w, h, MixColor(base, #VS_BG, 0.3))
  ;DrawRainbowH(0, 0, w, 2, 0.7)

  If FontUI : DrawingFont(FontID(FontUI)) : EndIf
  DrawingMode(#PB_2DDrawing_Transparent)

  DrawText(12, (h - TextHeight(StatusLeft$)) / 2, StatusLeft$, #VS_TEXT)

  If StatusRight$ <> ""
    tw = TextWidth(StatusRight$)
    DrawText(w - tw - 12, (h - TextHeight(StatusRight$)) / 2,
             StatusRight$, #VS_TEXT_DIM)
  EndIf

  StopDrawing()
EndProcedure

Procedure Status(text$, mode = 0)
  StatusLeft$ = text$
  StatusMode  = mode
  DrawStatusBar()
EndProcedure

Procedure UpdateCaretStatus()
  Protected pos, ln, col, total

  pos   = ScintillaSendMessage(#GAD_EDITOR, #SCI_GETCURRENTPOS, 0, 0)
  ln    = ScintillaSendMessage(#GAD_EDITOR, #SCI_LINEFROMPOSITION, pos, 0) + 1
  col   = pos - ScintillaSendMessage(#GAD_EDITOR, #SCI_POSITIONFROMLINE, ln - 1, 0) + 1
  total = ScintillaSendMessage(#GAD_EDITOR, #SCI_GETLINECOUNT, 0, 0)

  StatusRight$ = "Стр " + Str(ln) + " : Кол " + Str(col) +
                 "     Всего " + Str(total) + "     UTF-8"
  DrawStatusBar()
EndProcedure

;===============================================================================
; РАСКЛАДКА
;===============================================================================
Procedure LayoutUI()
  Protected w, h, splitH, pos, bx, gap = 6

  w = WindowWidth(0)
  h = WindowHeight(0)

  ;ResizeGadget(#GAD_TOOLBAR, 0, 0, w+700, #TOOLBAR_H+40)

  bx = 10
  ForEach TB()
    Select TB()\cmd
      Case #CMD_OPEN  : TB()\w = 80
      Case #CMD_SAVE  : TB()\w = 90
      Case #CMD_RUN   : TB()\w = 100
      Case #CMD_COMPILE : TB()\w = 110  ; Ширина для кнопки компиляции
      Case #CMD_CLEAR : TB()\w = 90
      Case #CMD_EXIT  : TB()\w = 70
      Default         : TB()\w = 80
    EndSelect
    TB()\h = 30
    TB()\y = (#TOOLBAR_H - #STRIPE_H - TB()\h) / 2

    If TB()\cmd = #CMD_EXIT
      TB()\x = w +640 - TB()\w - 10
    Else
      TB()\x = bx
      bx + TB()\w + gap
    EndIf
  Next

  splitH = h - #TOOLBAR_H - #STATUS_H
  If splitH < 140 : splitH = 140 : EndIf
 
 ; ResizeGadget(#GAD_SPLIT, 0, #TOOLBAR_H-20, w, splitH)

  pos = (splitH * #EDITOR_RATIO) / 100
  If pos < 60 : pos = 60 : EndIf
  If pos > splitH - 60 : pos = splitH - 60 : EndIf
  SetGadgetState(#GAD_SPLIT, pos)

  ;ResizeGadget(#GAD_STATUS, 0, h - #STATUS_H - 50, w, #STATUS_H+60)

  DrawToolbar()
  DrawStatusBar()
EndProcedure

;===============================================================================
; ЛОГ
;===============================================================================
Procedure.i LogLine(text$, kind = #LOG_INFO)
  Protected idx, prefix$

  Select kind
    Case #LOG_ERROR : prefix$ = "  ✖   "
    Case #LOG_WARN  : prefix$ = "  ▲   "
    Case #LOG_OK    : prefix$ = "  ✔   "
    Default         : prefix$ = "  ◆   "
  EndSelect

  AddGadgetItem(#GAD_LOG, -1, prefix$ + text$)
  idx = CountGadgetItems(#GAD_LOG) - 1
  SetGadgetState(#GAD_LOG, idx)
  SetGadgetState(#GAD_LOG, -1)
  ProcedureReturn idx
EndProcedure

Procedure.i ParseErrorLine(text$)
  Protected t$, p, i, c$, num$

  t$ = LCase(text$)

  p = FindString(t$, ".rainbow:")
  If p
    p + 9
  Else
    p = FindString(t$, "line ")
    If p
      p + 5
    Else
      p = FindString(t$, "строка ")
      If p
        p + 7
      Else
        p = FindString(t$, "стр. ")
        If p
          p + 5
        Else
          p = FindString(t$, "[")
          If p : p + 1 : EndIf
        EndIf
      EndIf
    EndIf
  EndIf

  If p = 0 : ProcedureReturn 0 : EndIf

  While p <= Len(t$) And Mid(t$, p, 1) = " " : p + 1 : Wend

  num$ = ""
  For i = p To Len(t$)
    c$ = Mid(t$, i, 1)
    If c$ >= "0" And c$ <= "9"
      num$ + c$
    Else
      Break
    EndIf
  Next

  If num$ = "" : ProcedureReturn 0 : EndIf
  ProcedureReturn Val(num$)
EndProcedure

Procedure MarkErrorLine(lineNo)
  Protected total, s, e

  total = ScintillaSendMessage(#GAD_EDITOR, #SCI_GETLINECOUNT, 0, 0)
  If lineNo < 1 Or lineNo > total : ProcedureReturn : EndIf

  ScintillaSendMessage(#GAD_EDITOR, #SCI_MARKERADD, lineNo - 1, #MARKER_ERROR)

  s = ScintillaSendMessage(#GAD_EDITOR, #SCI_POSITIONFROMLINE, lineNo - 1, 0)
  e = ScintillaSendMessage(#GAD_EDITOR, #SCI_GETLINEENDPOSITION, lineNo - 1, 0)
  If e > s
    ScintillaSendMessage(#GAD_EDITOR, #SCI_SETINDICATORCURRENT, #IND_ERROR, 0)
    ScintillaSendMessage(#GAD_EDITOR, #SCI_INDICATORFILLRANGE, s, e - s)
  EndIf
EndProcedure

Procedure ClearErrorMarks()
  Protected len

  ScintillaSendMessage(#GAD_EDITOR, #SCI_MARKERDELETEALL, #MARKER_ERROR, 0)
  len = ScintillaSendMessage(#GAD_EDITOR, #SCI_GETLENGTH, 0, 0)
  If len > 0
    ScintillaSendMessage(#GAD_EDITOR, #SCI_SETINDICATORCURRENT, #IND_ERROR, 0)
    ScintillaSendMessage(#GAD_EDITOR, #SCI_INDICATORCLEARRANGE, 0, len)
  EndIf
EndProcedure

Procedure LogError(text$)
  Protected idx, ln

  idx = LogLine(text$, #LOG_ERROR)
  ln  = ParseErrorLine(text$)
  If ln > 0
    LogLineTarget(Str(idx)) = ln
    MarkErrorLine(ln)
  EndIf
EndProcedure

Procedure GotoEditorLine(lineNo)
  Protected pos, total

  total = ScintillaSendMessage(#GAD_EDITOR, #SCI_GETLINECOUNT, 0, 0)
  If lineNo < 1 Or lineNo > total : ProcedureReturn : EndIf

  pos = ScintillaSendMessage(#GAD_EDITOR, #SCI_POSITIONFROMLINE, lineNo - 1, 0)
  ScintillaSendMessage(#GAD_EDITOR, #SCI_GOTOPOS, pos, 0)
  ScintillaSendMessage(#GAD_EDITOR, #SCI_ENSUREVISIBLEENFORCEPOLICY, lineNo - 1, 0)
  ScintillaSendMessage(#GAD_EDITOR, #SCI_SCROLLCARET, 0, 0)
  SetActiveGadget(#GAD_EDITOR)
EndProcedure

;===============================================================================
; Файлы
;===============================================================================
Procedure.i SaveEditorTo(File$)
  Protected f, len, *buf, ok = #False

  f = CreateFile(#PB_Any, File$)
  If Not f : ProcedureReturn #False : EndIf

  len  = ScintillaSendMessage(#GAD_EDITOR, #SCI_GETLENGTH, 0, 0)
  *buf = AllocateMemory(len + 1)
  If *buf
    ScintillaSendMessage(#GAD_EDITOR, #SCI_GETTEXT, len + 1, *buf)
    WriteData(f, *buf, len)
    FreeMemory(*buf)
    ok = #True
  EndIf
  CloseFile(f)
  ProcedureReturn ok
EndProcedure

Procedure LoadFileToEditor(File$)
  Protected FileNumber, fSize.q, *Buffer, BytesRead, Total.q, prev = -1, pc
  Protected ChunkSize = 131072

  FileNumber = ReadFile(#PB_Any, File$)
  If Not FileNumber
    LogError("Не удалось открыть файл: " + File$)
    ProcedureReturn
  EndIf

  StopAsync()
  WordIndStart = -1
  WordIndEnd   = -1
  ClearErrorMarks()
  ClearMap(LogLineTarget())
  ScintillaSendMessage(#GAD_EDITOR, #SCI_CLEARALL, 0, 0)
  fSize = Lof(FileNumber)

  If fSize > 0
    Status("Загрузка... (" + Str(fSize / 1024) + " KB)")

    *Buffer = AllocateMemory(ChunkSize)
    If *Buffer
      Repeat
        BytesRead = ReadData(FileNumber, *Buffer, ChunkSize)
        If BytesRead > 0
          ScintillaSendMessage(#GAD_EDITOR, #SCI_APPENDTEXT, BytesRead, *Buffer)
          Total + BytesRead
          pc = (Total * 100) / fSize
          If pc <> prev And pc % 5 = 0
            prev = pc
            Status("Загрузка... " + Str(pc) + "%")
            While WindowEvent() : Wend
          EndIf
        EndIf
      Until BytesRead <= 0 Or Total >= fSize
      FreeMemory(*Buffer)
    EndIf

    If fSize > #ASYNC_THRESHOLD
      StartAsyncHighlight(#GAD_EDITOR)
    Else
      Status("Подсветка...")
      HighlightText(#GAD_EDITOR)
      Status("Загружено: " + GetFilePart(File$), 2)
    EndIf

    CurrentFile$ = File$
    SetWindowTitle(0, "LGBTScript IDE v9.0  —  " + File$)
    LogLine("Загружен файл: " + File$ + " (" + Str(fSize) + " байт)", #LOG_OK)
  EndIf

  CloseFile(FileNumber)
EndProcedure

;===============================================================================
; КОМПИЛЯЦИЯ В .EXE
;===============================================================================
Procedure CompileToExe()
  Protected Process, out$, err$, lo$
  Protected errCount, warnCount, exitCode
  Protected startMs.q, elapsed.q
  Protected params$, workDir$, target$, exeName$

  ClearGadgetItems(#GAD_LOG)
  ClearMap(LogLineTarget())
  ClearErrorMarks()
  LastErrCount  = 0
  LastWarnCount = 0
  LastExitCode  = 0

  If CurrentFile$ <> ""
    target$ = CurrentFile$
  Else
    target$ = GetCurrentDirectory() + "main.rainbow"
  EndIf

  ; Сохраняем файл перед компиляцией
  If Not SaveEditorTo(target$)
    LogError("Не удалось сохранить " + target$)
    Status("Ошибка сохранения", 1)
    ProcedureReturn
  EndIf
  LogLine("Файл сохранён: " + GetFilePart(target$), #LOG_OK)

  ; Запрашиваем имя для .exe файла
  exeName$ = SaveFileRequester("Сохранить исполняемый файл", 
                               GetPathPart(target$) + GetFilePart(target$, #PB_FileSystem_NoExtension) + ".exe",
                               "EXE файлы (*.exe)|*.exe|Все файлы (*.*)|*.*", 0)
  
  If exeName$ = ""
    LogLine("Компиляция отменена")
    ProcedureReturn
  EndIf

  If FileSize("rb.exe") <= 0
    LogError("rb.exe не найден в директории: " + GetCurrentDirectory())
    LogLine("Поместите rb.exe рядом с IDE")
    Status("rb.exe не найден", 1)
    ProcedureReturn
  EndIf

  workDir$ = GetCurrentDirectory()
  params$  = "-b " + Chr(34) + target$ + Chr(34) + " " + Chr(34) + exeName$ + Chr(34)

  LogLine("Команда: rb.exe " + params$)
  LogLine("Каталог: " + workDir$)
  LogLine("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
  Status("Компиляция...")
  SetRunEnabled(#False)

  startMs = ElapsedMilliseconds()

  Process = RunProgram("rb.exe", params$, workDir$,
            #PB_Program_Open | #PB_Program_Read | #PB_Program_Error | #PB_Program_Hide)

  If Not Process
    CompilerIf #PB_Compiler_OS = #PB_OS_Windows
      LogError("Не удалось запустить rb.exe. Код Windows: " + Str(GetLastError_()))
    CompilerElse
      LogError("Не удалось запустить rb.exe")
    CompilerEndIf
    LogLine("Проверьте разрядность файла и наличие зависимых DLL")
    Status("Компиляция не удалась", 1)
    SetRunEnabled(#True)
    ProcedureReturn
  EndIf

  While ProgramRunning(Process)
    While AvailableProgramOutput(Process)
      out$ = ReadProgramString(Process)
      If out$ <> ""
        lo$ = LCase(out$)
        If FindString(lo$, "error") Or FindString(lo$, "ошибка") Or
           FindString(lo$, "fatal") Or FindString(lo$, "exception")
          LogError(out$)
          errCount + 1
        ElseIf FindString(lo$, "warning") Or FindString(lo$, "предупрежд")
          LogLine(out$, #LOG_WARN)
          warnCount + 1
        Else
          LogLine(out$)
        EndIf
      EndIf
    Wend

    err$ = ReadProgramError(Process)
    If err$ <> ""
      LogError(err$)
      errCount + 1
    EndIf

    While WindowEvent() : Wend
    Delay(1)
  Wend

  ; Читаем оставшиеся выводы
  While AvailableProgramOutput(Process)
    out$ = ReadProgramString(Process)
    If out$ <> ""
      lo$ = LCase(out$)
      If FindString(lo$, "error") Or FindString(lo$, "ошибка") Or
         FindString(lo$, "fatal") Or FindString(lo$, "exception")
        LogError(out$)
        errCount + 1
      ElseIf FindString(lo$, "warning") Or FindString(lo$, "предупрежд")
        LogLine(out$, #LOG_WARN)
        warnCount + 1
      Else
        LogLine(out$)
      EndIf
    EndIf
  Wend

  Repeat
    err$ = ReadProgramError(Process)
    If err$ = "" : Break : EndIf
    LogError(err$)
    errCount + 1
  ForEver

  exitCode = ProgramExitCode(Process)
  CloseProgram(Process)

  LastExitCode  = exitCode
  LastErrCount  = errCount
  LastWarnCount = warnCount

  elapsed = ElapsedMilliseconds() - startMs
  LogLine("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

  If exitCode = 0 And errCount = 0
    LogLine("✅ Компиляция успешно завершена за " + Str(elapsed) + " мс", #LOG_OK)
    LogLine("📁 Создан файл: " + exeName$, #LOG_OK)
    Status("✅ Компиляция завершена: " + GetFilePart(exeName$), 2)
    
    ; Предлагаем открыть папку с файлом
    If MessageRequester("Компиляция успешна", 
                        "Исполняемый файл создан:" + #CRLF$ + exeName$ + #CRLF$ + #CRLF$ + 
                        "Открыть папку с файлом?", #PB_MessageRequester_YesNo) = #PB_MessageRequester_Yes
      CompilerIf #PB_Compiler_OS = #PB_OS_Windows
        RunProgram("explorer.exe", "/select," + Chr(34) + exeName$ + Chr(34), "")
      CompilerElse
        RunProgram("open", GetPathPart(exeName$), "")
      CompilerEndIf
    EndIf
  Else
    LogError("❌ Компиляция завершилась с ошибками")
    LogError("Код возврата: " + Str(exitCode) +
             " | ошибок: " + Str(errCount) +
             " | предупреждений: " + Str(warnCount) +
             " | " + Str(elapsed) + " мс")
    
    Select exitCode
      Case 1   : LogLine("Код 1: общая ошибка компиляции")
      Case 2   : LogLine("Код 2: синтаксическая ошибка в коде")
      Case 3   : LogLine("Код 3: файл не найден или недоступен")
      Case 127 : LogLine("Код 127: команда не найдена")
    EndSelect

    Status("❌ Ошибок: " + Str(errCount) + "   код " + Str(exitCode), 1)
  EndIf

  SetRunEnabled(#True)
EndProcedure

;===============================================================================
; Запуск rb.exe
;===============================================================================
Procedure RunRainbow()
  Protected Process, out$, err$, lo$
  Protected errCount, warnCount, exitCode
  Protected startMs.q, elapsed.q
  Protected params$, workDir$, target$

  ClearGadgetItems(#GAD_LOG)
  ClearMap(LogLineTarget())
  ClearErrorMarks()
  LastErrCount  = 0
  LastWarnCount = 0
  LastExitCode  = 0

  If CurrentFile$ <> ""
    target$ = CurrentFile$
  Else
    target$ = GetCurrentDirectory() + "main.rainbow"
  EndIf

  If Not SaveEditorTo(target$)
    LogError("Не удалось сохранить " + target$)
    Status("Ошибка сохранения", 1)
    ProcedureReturn
  EndIf
  LogLine("Файл сохранён: " + GetFilePart(target$), #LOG_OK)

  If FileSize("rb.exe") <= 0
    LogError("rb.exe не найден в директории: " + GetCurrentDirectory())
    LogLine("Поместите rb.exe рядом с IDE")
    Status("rb.exe не найден", 1)
    ProcedureReturn
  EndIf

  workDir$ = GetCurrentDirectory()
  params$  = "-lgbt " + Chr(34) + target$ + Chr(34)

  LogLine("Команда: rb.exe " + params$)
  LogLine("Каталог: " + workDir$)
  LogLine("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
  Status("Выполняется...")
  SetRunEnabled(#False)

  startMs = ElapsedMilliseconds()

  Process = RunProgram("rb.exe", params$, workDir$,
            #PB_Program_Open | #PB_Program_Read | #PB_Program_Error | #PB_Program_Hide)

  If Not Process
    CompilerIf #PB_Compiler_OS = #PB_OS_Windows
      LogError("Не удалось запустить rb.exe. Код Windows: " + Str(GetLastError_()))
    CompilerElse
      LogError("Не удалось запустить rb.exe")
    CompilerEndIf
    LogLine("Проверьте разрядность файла и наличие зависимых DLL")
    Status("Запуск не удался", 1)
    SetRunEnabled(#True)
    ProcedureReturn
  EndIf

  While ProgramRunning(Process)
    While AvailableProgramOutput(Process)
      out$ = ReadProgramString(Process)
      If out$ <> ""
        lo$ = LCase(out$)
        If FindString(lo$, "error") Or FindString(lo$, "ошибка") Or
           FindString(lo$, "fatal") Or FindString(lo$, "exception")
          LogError(out$)
          errCount + 1
        ElseIf FindString(lo$, "warning") Or FindString(lo$, "предупрежд")
          LogLine(out$, #LOG_WARN)
          warnCount + 1
        Else
          LogLine(out$)
        EndIf
      EndIf
    Wend

    err$ = ReadProgramError(Process)
    If err$ <> ""
      LogError(err$)
      errCount + 1
    EndIf

    While WindowEvent() : Wend
    Delay(1)
  Wend

  While AvailableProgramOutput(Process)
    out$ = ReadProgramString(Process)
    If out$ <> ""
      lo$ = LCase(out$)
      If FindString(lo$, "error") Or FindString(lo$, "ошибка") Or
         FindString(lo$, "fatal") Or FindString(lo$, "exception")
        LogError(out$)
        errCount + 1
      ElseIf FindString(lo$, "warning") Or FindString(lo$, "предупрежд")
        LogLine(out$, #LOG_WARN)
        warnCount + 1
      Else
        LogLine(out$)
      EndIf
    EndIf
  Wend

  Repeat
    err$ = ReadProgramError(Process)
    If err$ = "" : Break : EndIf
    LogError(err$)
    errCount + 1
  ForEver

  exitCode = ProgramExitCode(Process)
  CloseProgram(Process)

  LastExitCode  = exitCode
  LastErrCount  = errCount
  LastWarnCount = warnCount

  elapsed = ElapsedMilliseconds() - startMs
  LogLine("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

  If exitCode = 0 And errCount = 0
    LogLine("Завершено успешно за " + Str(elapsed) + " мс (предупреждений: " +
            Str(warnCount) + ")", #LOG_OK)
    Status("Выполнено за " + Str(elapsed) + " мс", 2)
  Else
    LogError("Код возврата: " + Str(exitCode) +
             " | ошибок: " + Str(errCount) +
             " | предупреждений: " + Str(warnCount) +
             " | " + Str(elapsed) + " мс")
    
    SetGadgetState(#GAD_LOG, CountGadgetItems(#GAD_LOG))
    
    Select exitCode
      Case 1   : LogLine("Код 1: общая ошибка выполнения скрипта")
      Case 2   : LogLine("Код 2: синтаксическая ошибка / неверные аргументы")
      Case 3   : LogLine("Код 3: файл не найден или недоступен")
      Case 127 : LogLine("Код 127: команда не найдена")
      Case -1073741819
                 LogLine("Код 0xC0000005: нарушение доступа (crash интерпретатора)")
      Case -1073741510
                 LogLine("Код 0xC000013A: прервано пользователем (Ctrl+C)")
    EndSelect

    Status("Ошибок: " + Str(errCount) + "   код " + Str(exitCode), 1)
  EndIf

  If errCount > 0
    LogLine("Двойной клик по строке ✖ — переход к строке кода")
  EndIf

  SetRunEnabled(#True)
EndProcedure

;===============================================================================
; Обработка команд
;===============================================================================
Procedure DoCommand(cmd)
  Protected File$

  Select cmd

    Case 1919
      SetGadgetText(#GAD_EDITOR, "")
      CurrentFile$=""
      SetWindowTitle(0, "LGBTScript IDE v9.0")
   
    Case #CMD_OPEN
      File$ = OpenFileRequester("Открыть файл", "",
              "Rainbow files (*.rainbow)|*.rainbow|All files (*.*)|*.*", 0)
      If File$ : LoadFileToEditor(File$) : EndIf

    Case #CMD_SAVE
      If CurrentFile$ = ""
        File$ = SaveFileRequester("Сохранить файл", "main.rainbow",
                "Rainbow files (*.rainbow)|*.rainbow|All files (*.*)|*.*", 0)
      Else
        File$ = CurrentFile$
      EndIf
      If File$
        If SaveEditorTo(File$)
          CurrentFile$ = File$
          SetWindowTitle(0, "LGBTScript IDE v9.0  —  " + File$)
          LogLine("Файл сохранён: " + GetFilePart(File$), #LOG_OK)
          Status("Сохранено: " + GetFilePart(File$), 2)
        Else
          LogError("Не удалось сохранить: " + File$)
          Status("Ошибка сохранения", 1)
        EndIf
      EndIf

    Case #CMD_SAVEAS
      File$ = SaveFileRequester("Сохранить как", "main.rainbow",
              "Rainbow files (*.rainbow)|*.rainbow|All files (*.*)|*.*", 0)
      If File$
        If SaveEditorTo(File$)
          CurrentFile$ = File$
          SetWindowTitle(0, "LGBTScript IDE v9.0  —  " + File$)
          LogLine("Файл сохранён: " + GetFilePart(File$), #LOG_OK)
          Status("Сохранено: " + GetFilePart(File$), 2)
        Else
          LogError("Не удалось создать файл: " + File$)
        EndIf
      EndIf

    Case #CMD_EXIT  : Quit = 1
    Case #CMD_CUT   : ScintillaSendMessage(#GAD_EDITOR, #SCI_CUT,   0, 0)
    Case #CMD_COPY  : ScintillaSendMessage(#GAD_EDITOR, #SCI_COPY,  0, 0)
    Case #CMD_PASTE : ScintillaSendMessage(#GAD_EDITOR, #SCI_PASTE, 0, 0)
    Case #CMD_UNDO  : ScintillaSendMessage(#GAD_EDITOR, #SCI_UNDO,  0, 0)
    Case #CMD_REDO  : ScintillaSendMessage(#GAD_EDITOR, #SCI_REDO,  0, 0)
    Case #CMD_RUN   : RunRainbow()
    Case #CMD_COMPILE : CompileToExe()  ; Обработка команды компиляции

    Case #CMD_CLEAR
      ClearGadgetItems(#GAD_LOG)
      ClearMap(LogLineTarget())
      ClearErrorMarks()
      Status("Лог очищен")

    Case #CMD_ABOUT
      MessageRequester("О программе",
        "LGBTScript IDE v9.0")
  EndSelect
EndProcedure

;===============================================================================
; ГЛАВНОЕ ОКНО
;===============================================================================
Define Ev, sel, tgt, mx, my, cmd, *initText

If OpenWindow(0, 0, 0, 1200, 480, "LGBTScript IDE v9.0",
              #PB_Window_ScreenCentered | 
             #PB_Window_Maximize|#PB_Window_BorderLess)

  SetWindowColor(0, #VS_BG)
  CompilerIf #PB_Compiler_OS = #PB_OS_Windows
    StylizeTitleBar(0)
  CompilerEndIf

  InitKeywords()
  InitBuiltinFunctions()
  InitToolbar()

  FontUI     = LoadFont(#PB_Any, UIFont$, 9)
  FontUIBold = LoadFont(#PB_Any, UIFont$, 9, #PB_Font_Bold)
  FontMono   = LoadFont(#PB_Any, FontName$, 11)

  ;--- меню ---
  If CreateMenu(0, WindowID(0))
    MenuTitle("Файл")
      MenuItem(1919,   "Создать" + Chr(9) + "Ctrl+N")
      MenuItem(#CMD_OPEN,   "Открыть..." + Chr(9) + "Ctrl+O")
      MenuItem(#CMD_SAVE,   "Сохранить" + Chr(9) + "Ctrl+S")
      MenuItem(#CMD_SAVEAS, "Сохранить как...")
      MenuBar()
      MenuItem(#CMD_EXIT,   "Выход")
    MenuTitle("Правка")
      MenuItem(#CMD_CUT,   "Вырезать" + Chr(9) + "Ctrl+X")
      MenuItem(#CMD_COPY,  "Копировать" + Chr(9) + "Ctrl+C")
      MenuItem(#CMD_PASTE, "Вставить" + Chr(9) + "Ctrl+V")
      MenuBar()
      MenuItem(#CMD_UNDO,  "Отменить" + Chr(9) + "Ctrl+Z")
      MenuItem(#CMD_REDO,  "Повторить" + Chr(9) + "Ctrl+Y")
    MenuTitle("Запуск")
      MenuItem(#CMD_RUN,   "Выполнить" + Chr(9) + "F5")
      MenuItem(#CMD_COMPILE, "Создать исполняемый файл")  ; Добавлен пункт меню
     
      MenuItem(#CMD_CLEAR, "Очистить лог")
    MenuTitle("Справка")
      MenuItem(#CMD_ABOUT, "О программе")
  EndIf

  AddKeyboardShortcut(0, #PB_Shortcut_F5, #CMD_RUN)
  AddKeyboardShortcut(0, #PB_Shortcut_Control | #PB_Shortcut_O, #CMD_OPEN)
  AddKeyboardShortcut(0, #PB_Shortcut_Control | #PB_Shortcut_S, #CMD_SAVE)
  AddKeyboardShortcut(0, #PB_Shortcut_Control | #PB_Shortcut_N, 1919)
  AddKeyboardShortcut(0, #PB_Shortcut_Escape, #CMD_RUN)

  ;--- тулбар ---
  CanvasGadget(#GAD_TOOLBAR, 0, 0, 2500, #TOOLBAR_H+0, #PB_Canvas_Keyboard)

  ;--- редактор и лог ---
  ScintillaGadget(#GAD_EDITOR, 0, 0, 10, 10, @ScintillaCB())
  SetupScintilla(#GAD_EDITOR)

  ListViewGadget(#GAD_LOG, 0, 0, 10, 10)
  SetGadgetColor(#GAD_LOG, #PB_Gadget_BackColor,  #VS_BG)
  SetGadgetColor(#GAD_LOG, #PB_Gadget_FrontColor, #VS_TEXT)
  If FontMono : SetGadgetFont(#GAD_LOG, FontID(FontMono)) : EndIf
  GadgetToolTip(#GAD_LOG, "Двойной клик по ошибке — переход к строке кода")

  ;--- сплиттер ---
  SplitterGadget(#GAD_SPLIT, 0, #TOOLBAR_H+30, 1280,
                 670 - #TOOLBAR_H - #STATUS_H+20,
                 #GAD_EDITOR, #GAD_LOG)
  SetGadgetColor(#GAD_SPLIT, #PB_Gadget_BackColor, #PR_VIOLET)

  ;--- статус ---
  CanvasGadget(#GAD_STATUS, 0, 780 - #STATUS_H+30, 1200, #STATUS_H)

  LayoutUI()
  BindEvent(#PB_Event_Timer, @TimerHandler(), 0)

  ;--- стартовый текст ---
  *initText = UTF8("@ LGBTScript IDE" + Chr(10) +
                   "" + Chr(10) +
                   "gay alex = 7;" + Chr(10) + 
                   "lesbian masha = " + Chr(34) + "лесбиянок" + Chr(34) + ";" + Chr(10) + 
                   "comingout alex + " + Chr(34) + " " + Chr(34) + " + masha;")
  ScintillaSendMessage(#GAD_EDITOR, #SCI_SETTEXT, 0, *initText)
  FreeMemory(*initText)

  HighlightText(#GAD_EDITOR)
  UpdateCaretStatus()
  LogLine("LGBTScript IDE v9.0 запущен", #LOG_OK)
 
  ;===========================================================================
  ; ЦИКЛ СОБЫТИЙ
  ;===========================================================================
  Repeat
    Ev = WaitWindowEvent()

    Select Ev
      
        
      Case #PB_Event_Menu
        DoCommand(EventMenu())
      
      Case #PB_Event_Gadget
        Select EventGadget()
         
            
            
          Case #GAD_TOOLBAR
            mx = GetGadgetAttribute(#GAD_TOOLBAR, #PB_Canvas_MouseX)
            my = GetGadgetAttribute(#GAD_TOOLBAR, #PB_Canvas_MouseY)

            Select EventType()
              Case #PB_EventType_MouseMove
                ;ToolbarHover(mx, my)

              Case #PB_EventType_MouseLeave
                ForEach TB()
                  TB()\hover = #False
                  TB()\down  = #False
                Next
              ;  DrawToolbar()

              Case #PB_EventType_LeftButtonDown
                ForEach TB()
                  If mx >= TB()\x And mx < TB()\x + TB()\w And
                     my >= TB()\y And my < TB()\y + TB()\h
                    TB()\down = #True
                  EndIf
                Next
               ; DrawToolbar()

              Case #PB_EventType_LeftButtonUp
                cmd = ToolbarHit(mx, my)
                ForEach TB()
                  TB()\down = #False
                Next
                ;DrawToolbar()
                If cmd
                  ForEach TB()
                    If TB()\cmd = cmd And TB()\disabled : cmd = 0 : EndIf
                  Next
                EndIf
                If cmd : DoCommand(cmd) : EndIf
            EndSelect

          Case #GAD_EDITOR
            If EventType() = #PB_EventType_Change
              ClearErrorMarks()
            EndIf
            
             
          Case #GAD_LOG
            If EventType() = #PB_EventType_LeftDoubleClick
              sel = GetGadgetState(#GAD_LOG)
              If sel >= 0 And FindMapElement(LogLineTarget(), Str(sel))
                tgt = LogLineTarget()
                GotoEditorLine(tgt)
                Status("Переход к строке " + Str(tgt))
              EndIf
            EndIf
            
            
            If EventType() = #PB_EventType_RightClick
              MessageRequester("LGBTScript IDE v9.0", "Текст скопирован в буфер обмена")
           SetClipboardText(GetGadgetItemText(#GAD_LOG, GetGadgetState(#GAD_LOG)))
            EndIf

        EndSelect

      Case #PB_Event_SizeWindow
        LayoutUI()

    EndSelect

  Until Ev = #PB_Event_CloseWindow Or Quit = 1

  StopAsync()
  If *TextBuf : FreeMemory(*TextBuf) : EndIf
EndIf
; IDE Options = PureBasic 6.21 (Windows - x64)
; CursorPosition = 1754
; FirstLine = 1747
; Folding = --------------
; EnableXP
; DPIAware
; UseIcon = ico.ico
; Executable = LGBTScript IDE.exe