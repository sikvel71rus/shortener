// Package main implements staticlint, a project multichecker based on
// golang.org/x/tools/go/analysis.
//
// Run it from the repository root:
//
//	go run ./cmd/staticlint ./...
//
// For repeated local use it can also be built and then launched as a normal
// analyzer binary:
//
//	go build -o staticlint ./cmd/staticlint
//	./staticlint ./...
//
// The command accepts the standard multichecker flags. For example,
// -help prints the full list of enabled analyzers and flags, and -NAME=false
// disables a particular analyzer for one run.
//
// Enabled analyzer groups:
//
//   - Standard analyzers from golang.org/x/tools/go/analysis/passes:
//     appends checks suspicious append calls; asmdecl validates assembly
//     declarations; assign detects useless assignments; atomic detects common
//     sync/atomic mistakes; atomicalign checks 64-bit atomic alignment; bools
//     finds redundant or impossible boolean expressions; buildtag validates
//     build constraints; cgocall detects invalid cgo pointer passing;
//     composite reports unkeyed composite literals from external packages;
//     copylock reports copying values that contain locks; deepequalerrors
//     warns about reflect.DeepEqual on error values; defers detects misuse of
//     defer; directive validates Go toolchain comments; errorsas checks
//     errors.As targets; framepointer validates frame pointer assembly;
//     hostport checks net.JoinHostPort candidates; httpmux checks invalid
//     net/http patterns; httpresponse finds HTTP response handling mistakes;
//     ifaceassert detects impossible interface assertions; loopclosure reports
//     captured loop variables; lostcancel reports discarded context.CancelFunc
//     values; nilfunc detects useless function nil comparisons; printf
//     validates Printf-style calls; reflectvaluecompare detects invalid
//     comparisons involving reflect.Value; shift checks invalid shifts;
//     sigchanyzer reports unbuffered signal channels; slog validates structured
//     logging calls; sortslice checks sort.Slice comparisons; stdmethods checks
//     signatures of standard methods; stdversion checks symbols against the
//     configured Go version; stringintconv warns about integer-to-string
//     conversions; structtag validates struct tags; testinggoroutine detects
//     fatal test calls from goroutines; tests checks common test mistakes;
//     timeformat validates time layouts; unmarshal validates unmarshalling
//     targets; unreachable finds unreachable code; unsafeptr detects invalid
//     uintptr/unsafe.Pointer conversions; unusedresult reports ignored results
//     of selected functions; unusedwrite detects writes that are never read;
//     waitgroup detects sync.WaitGroup misuse.
//   - All SA analyzers from Staticcheck. The SA class focuses on likely bugs:
//     invalid API usage, dead branches, impossible comparisons, broken
//     synchronization, nil dereferences, ineffective assignments, invalid
//     regular expressions, suspicious time and template code, and other
//     correctness defects.
//   - Staticcheck S1000. This analyzer belongs to the S class and suggests
//     simpler forms for selected code constructs.
//   - bodyclose. This public analyzer checks that HTTP response bodies are
//     closed, preventing resource leaks.
//   - rowserrcheck. This public analyzer checks that database/sql Rows.Err is
//     inspected after rows iteration.
//   - osexitcheck. This project analyzer forbids direct os.Exit calls inside
//     func main of package main, so shutdown and deferred cleanup remain
//     observable and testable.
package main
