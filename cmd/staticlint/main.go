// staticlint является утилитой, предназначенной для выполнения статического анализа Go-кода,
// и представляет собой объединение нескольких статических анализаторов.
//
// В состастав staticlint входят следующие статические анализаторы.
// - contextcheck - проверяет, использует ли функция ненаследуемый контекст.
// - bodyclose - проверяет, успешно ли закрыто тело ответа HTTP.
// - assign - обнаруживает бесполезные присвоения.
// - atomic - проверяет распространенные ошибки с помощью пакета sync/atomic.
// - atomicalign - проверяет аргументы, не выровненные по 64 битам, для функций синхронизации/атомарных функций.
// - bools - обнаруживает распространенные ошибки, связанные с логическими операторами.
// - buildtag - проверяет корректность тегов сборки.
// - composite - проверяет наличие составных литералов без ключей.
// - copylock - проверяет наличие блокировок, ошибочно переданных по значению.
// - deepequalerrors - проверяет использование Reflection.DeepEqual со значениями ошибок.
// - defers - проверяет распространенные ошибки в операторах defer.
// - errorsas - проверяет, что второй аргумент error.As является указателем на ошибку реализации типа.
// - httpresponse - проверяет наличие ошибок обработки HTTP-ответов.
// - ifaceassert - помечает невозможные утверждения типа интерфейса.
// - loopclosure - проверяет наличие ссылок на переменные цикла внутри вложенных функций.
// - lostcancel - проверяет отсутствие вызова функции отмены контекста.
// - nilfunc - проверяет бесполезные сравнения с nil.
// - printf - проверяет согласованность строк и аргументов формата Printf.
// - shadow - проверяет наличие затененных переменных.
// - sigchanyzer - обнаруживает неправильное использование небуферизованного сигнала в качестве аргумента signal.Notify.
// - sortslice - проверяет вызовы sort.Slice, которые не используют тип слайса в качестве первого аргумента.
// - stdmethods - проверяет наличие орфографических ошибок в сигнатурах методов, аналогичных общеизвестным интерфейсам.
// - stringintconv - помечает преобразования типов из целых чисел в строки вида string(x), где x - целое число.
// - unmarshal - проверяет передачу типов, не являющихся указателями или
// неинтерфейсами, для функций демаршалинга и декодирования.
// - unreachable - проверяет наличие недостижимого кода.
// - unusedresult - проверяет неиспользуемые результаты вызовов некоторых функций.
// - unusedwrite - проверяет наличие неиспользуемых записей в элементы объекта структуры или массива.
// - staticcheck - содержит множество анализаторов, позволяющих обнаружить ошибки и проблемы с производительностью.
// - stylecheck - содержит множество анализаторов, обеспечивающих соблюдение правил стиля.
// - simple - содержит множество анализаторов, которые призваны упростить код.
// - quickfix - содержит множество анализаторов, реализующие рефакторинг кода.
// - exitmaincheck - проверяет, используется ли функция os.Exit() в функции main() пакета main.
//
// Запуск полной проверки: staticlit [package].
//
// Запуск конкретного статического анализатора: staticlit [analyser] [package].
//
// Справка: staticlint -h.
package main

import (
	"github.com/KryukovO/metricscollector/internal/exitmaincheck"
	"github.com/kkHAIKE/contextcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	mychecks := []*analysis.Analyzer{
		// Некоторые анализаторы golang.org/x/tools/go/analysis/passes
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		deepequalerrors.Analyzer,
		defers.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,

		// Анализатор contextcheck
		contextcheck.NewAnalyzer(
			contextcheck.Configuration{
				DisableFact: true,
			},
		),

		// Анализатор bodyclose
		bodyclose.Analyzer,

		// Анализатор для поиска os.Exit
		exitmaincheck.Analyzer,
	}

	// Анализаторы пакета staticcheck.io
	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	for _, v := range stylecheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	for _, v := range simple.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	for _, v := range quickfix.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	multichecker.Main(
		mychecks...,
	)
}
