package linereader_test

import (
	"bytes"
	"io"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/asymmetric-research/go-commons/io/linereader"
	"github.com/asymmetric-research/go-commons/io/readchunkdump"
	"github.com/stretchr/testify/require"

	gocmd "github.com/go-cmd/cmd"
	hashiline "github.com/mitchellh/go-linereader"
)

func TestLineReader(t *testing.T) {
	expectedLines := strings.Split(report, "\n")

	r := linereader.New(strings.NewReader(report), 4096)
	var linesback [8192]byte
	var line []byte

	var err error
	var n int

	for err != io.EOF {
		n, _, err = r.ReadExtra(linesback[:])
		if n == 0 && err == io.EOF {
			continue
		}
		line = linesback[:n]
		require.Equal(t, expectedLines[0], string(line))
		expectedLines = expectedLines[1:]
	}

	require.ErrorIs(t, err, io.EOF)
	require.Emptyf(t, expectedLines, "should have produced as many lines as expected")
}

func TestLinesOfReaderTruncation(t *testing.T) {
	expectedLines := strings.Split(report, "\n")

	r := linereader.New(strings.NewReader(report), 4096)
	var linesback [10]byte
	var line []byte

	var err error
	var n int

	var i = 0
	for err != io.EOF {
		i += 1
		n, _, err = r.ReadExtra(linesback[:])
		if n == 0 && err == io.EOF {
			break
		}

		if err != nil {
			break
		}

		line = linesback[:n]
		if !strings.HasPrefix(expectedLines[0], string(line)) {
			t.Fatalf(
				"line does not have expected prefix:\n-line: %s\n-prefix: %s\n",
				expectedLines[0],
				string(line),
			)
		}
		expectedLines = expectedLines[1:]
	}

	require.ErrorIs(t, err, io.EOF)
	require.Emptyf(t, expectedLines, "should have produced as many lines as expected")
}

func TestReplay(t *testing.T) {
	_, currentFile, _, _ := runtime.Caller(0)
	currentDir := path.Dir(currentFile)

	r, err := readchunkdump.NewReplayer(
		path.Join(currentDir, "readerchunks0"),
	)
	require.NoError(t, err)
	lr := linereader.New(r, 1024*4)        // 4K read buffer
	backingBuf := [20 * 1024 * 1024]byte{} // 20MB max line

	for i := 0; ; i++ {
		n, dis, rerr := lr.ReadExtra(backingBuf[:])
		_ = dis

		rb := backingBuf[:n]

		if bytes.ContainsRune(rb, '\x00') {
			t.FailNow()
		}
		if rerr == io.EOF {
			return
		}
	}
}

// Unbuffered Benchmarks
func BenchmarkLineReaderUnbuffered(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			reader := NewLineByLineReader(report)
			runOurs(b, reader)
		}
	})
}

func BenchmarkHashicorpsUnbuffered(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			reader := NewLineByLineReader(report)
			runHashicorps(b, reader)
		}
	})
}

func BenchmarkGoCmdUnbuffered(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			reader := NewLineByLineReader(report)
			runGoCmds(b, reader)
		}
	})
}

// Buffered benchmarks
func BenchmarkLineReaderLargeReads(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			reader := strings.NewReader(report)
			runOurs(b, reader)
		}
	})
}

func BenchmarkHashicorpsLargeReads(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			reader := strings.NewReader(report)
			runHashicorps(b, reader)
		}
	})
}

func BenchmarkGoCmdLargeReads(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			reader := strings.NewReader(report)
			runGoCmds(b, reader)
		}
	})
}

func runOurs(t require.TestingT, r io.Reader) {
	var err error
	rd := linereader.T{}
	lineBacking := [8192]byte{}
	linereader.NewInto(&rd, r, 4096)

	cnt := 0
	for err == nil {
		_, _, err = rd.ReadExtra(lineBacking[:])
		cnt += 1
	}
	cnt -= 1 // account for the last error
	require.Equal(t, reportLineCount, cnt)
}

func runHashicorps(t require.TestingT, r io.Reader) {
	rd := hashiline.New(r)
	cnt := 0
	go rd.Run()
	for range rd.Ch {
		cnt += 1
	}
	require.Equal(t, 283, cnt)
}

func runGoCmds(t require.TestingT, r io.Reader) {
	streamchan := make(chan string)
	stream := gocmd.NewOutputStream(streamchan)

	go func() {

		// using io.Copy here would be cheating as it uses a large backing buffer and could get the file in one run. Lines rarely come in one go.
		// io.Copy(stream, reader)

		buf := make([]byte, 4096)

		for {
			n, _ := r.Read(buf)
			if n == 0 {
				break
			}
			r := buf[:n]
			for len(r) > 0 {
				wn, _ := stream.Write(r)
				r = r[wn:]
			}
		}

		stream.Flush()
		close(streamchan)
	}()

	cnt := 0
	for range stream.Lines() {
		cnt += 1
	}
	require.Equal(t, 283, cnt)
}

const report = `AFLTriage v1.0.0 by Grant Hernandez
[+] GDB is working (GNU gdb (Ubuntu 12.1-0ubuntu1~22.04.2) 12.1 - Python 3.10.12 (main, Jul 29 2024, 16:56:48) [GCC 11.4.0])
[+] Image triage cmdline: ./test_assets/ezbug_x86_64 @@
[+] Will output rawjson reports to terminal
[+] Triaging single ./test_assets/crashes/crash-1036e40820c11936e0b8d3069623cbecad6b6b95
[+] Triage timeout set to 90000ms
[+] Profiling target...
[+] Target profile: time=233.509096ms, mem=1KB
[+] Debugged profile: t=376.641514ms (1.61x), mem=36212KB (36212.00x)
[+] System memory available: 18865564 KB
[+] System cores available: 16
[+] Triaging 1 testcases
[+] Using 1 threads to triage
[+] Processing initial 1 test cases
[+] ./test_assets/crashes/crash-1036e40820c11936e0b8d3069623cbecad6b6b95: CRASH detected in LLVMFuzzerTestOneInput due to a fault at or near 0x0000555555693755 leading to SIGILL (si_signo=4) / ILL_ILLOPN (si_code=2)
[+] --- RAWJSON REPORT BEGIN ---
{
  "bucket": {
    "inputs": [
      "/home/nick/src/github.com/FuzzCorp/test_assets/ezbug_x86_64 (.text)+0x10ef05",
      "/home/nick/src/github.com/FuzzCorp/test_assets/ezbug_x86_64 (.text)+0x1afd1",
      "/home/nick/src/github.com/FuzzCorp/test_assets/ezbug_x86_64 (.text)+0x5123",
      "/home/nick/src/github.com/FuzzCorp/test_assets/ezbug_x86_64 (.text)+0xab67",
      "/home/nick/src/github.com/FuzzCorp/test_assets/ezbug_x86_64 (.text)+0x33cd3"
    ],
    "strategy": "afltriage",
    "strategy_result": "b6ddbbeed2004ab4e3095bc588be36a0"
  },
  "command_line": [
    "./test_assets/ezbug_x86_64",
    "@@"
  ],
  "debugger": "gdb",
  "report": {
    "child": {
      "stderr": "INFO: Running with entropic power schedule (0xFF, 100).\nINFO: Seed: 4118876947\nINFO: Loaded 1 modules   (8 inline 8-bit counters): 8 [0x5555556d9e88, 0x5555556d9e90), \nINFO: Loaded 1 PC tables (8 PCs): 8 [0x5555556d9e90,0x5555556d9f10), \n/home/nick/src/github.com/FuzzCorp/test_assets/ezbug_x86_64: Running 1 inputs 1 time(s) each.\nRunning: ./test_assets/crashes/crash-1036e40820c11936e0b8d3069623cbecad6b6b95\n",
      "stdout": "Warning: 'set logging on', an alias for the command 'set logging enabled', is deprecated.\nUse 'set logging enabled on'.\n\n"
    },
    "response": {
      "context": {
        "arch_info": {
          "address_bits": 64,
          "architecture": "i386:x86-64"
        },
        "primary_thread": {
          "backtrace": [
            {
              "address": 93824993539925,
              "module": "/home/nick/src/github.com/FuzzCorp/test_assets/ezbug_x86_64",
              "module_address": "/home/nick/src/github.com/FuzzCorp/test_assets/ezbug_x86_64 (.text)+0x10ef05",
              "relative_address": 1109765,
              "symbol": {
                "function_name": "LLVMFuzzerTestOneInput"
              }
            },
            {
              "address": 93824992540705,
              "module": "/home/nick/src/github.com/FuzzCorp/test_assets/ezbug_x86_64",
              "module_address": "/home/nick/src/github.com/FuzzCorp/test_assets/ezbug_x86_64 (.text)+0x1afd1",
              "relative_address": 110545,
              "symbol": {
                "function_name": "fuzzer::Fuzzer::ExecuteCallback(unsigned char const*, unsigned long)"
              }
            },
            {
              "address": 93824992450931,
              "module": "/home/nick/src/github.com/FuzzCorp/test_assets/ezbug_x86_64",
              "module_address": "/home/nick/src/github.com/FuzzCorp/test_assets/ezbug_x86_64 (.text)+0x5123",
              "relative_address": 20771,
              "symbol": {
                "function_name": "fuzzer::RunOneTest(fuzzer::Fuzzer*, char const*, unsigned long)"
              }
            },
            {
              "address": 93824992474039,
              "module": "/home/nick/src/github.com/FuzzCorp/test_assets/ezbug_x86_64",
              "module_address": "/home/nick/src/github.com/FuzzCorp/test_assets/ezbug_x86_64 (.text)+0xab67",
              "relative_address": 43879,
              "symbol": {
                "function_name": "fuzzer::FuzzerDriver(int*, char***, int (*)(unsigned char const*, unsigned long))"
              }
            },
            {
              "address": 93824992642339,
              "module": "/home/nick/src/github.com/FuzzCorp/test_assets/ezbug_x86_64",
              "module_address": "/home/nick/src/github.com/FuzzCorp/test_assets/ezbug_x86_64 (.text)+0x33cd3",
              "relative_address": 212179,
              "symbol": {
                "function_name": "main"
              }
            }
          ],
          "current_instruction": "ud2",
          "registers": [
            {
              "name": "rax",
              "pretty_value": "140737343975424",
              "size": 8,
              "type": "int64_t",
              "value": 140737343975424
            },
            {
              "name": "rbx",
              "pretty_value": "89678917140608",
              "size": 8,
              "type": "int64_t",
              "value": 89678917140608
            },
            {
              "name": "rcx",
              "pretty_value": "140737316352000",
              "size": 8,
              "type": "int64_t",
              "value": 140737316352000
            },
            {
              "name": "rdx",
              "pretty_value": "140737350028000",
              "size": 8,
              "type": "int64_t",
              "value": 140737350028000
            },
            {
              "name": "rsi",
              "pretty_value": "0",
              "size": 8,
              "type": "int64_t",
              "value": 0
            },
            {
              "name": "rdi",
              "pretty_value": "140737316352000",
              "size": 8,
              "type": "int64_t",
              "value": 140737316352000
            },
            {
              "name": "rbp",
              "pretty_value": "0x7fffffffdd10",
              "size": 8,
              "type": "",
              "value": 140737488346384
            },
            {
              "name": "rsp",
              "pretty_value": "0x7fffffffdca0",
              "size": 8,
              "type": "",
              "value": 140737488346272
            },
            {
              "name": "r8",
              "pretty_value": "140737353826304",
              "size": 8,
              "type": "int64_t",
              "value": 140737353826304
            },
            {
              "name": "r9",
              "pretty_value": "0",
              "size": 8,
              "type": "int64_t",
              "value": 0
            },
            {
              "name": "r10",
              "pretty_value": "128",
              "size": 8,
              "type": "int64_t",
              "value": 128
            },
            {
              "name": "r11",
              "pretty_value": "514",
              "size": 8,
              "type": "int64_t",
              "value": 514
            },
            {
              "name": "r12",
              "pretty_value": "88098369175568",
              "size": 8,
              "type": "int64_t",
              "value": 88098369175568
            },
            {
              "name": "r13",
              "pretty_value": "93824993829888",
              "size": 8,
              "type": "int64_t",
              "value": 93824993829888
            },
            {
              "name": "r14",
              "pretty_value": "88098369175600",
              "size": 8,
              "type": "int64_t",
              "value": 88098369175600
            },
            {
              "name": "r15",
              "pretty_value": "7",
              "size": 8,
              "type": "int64_t",
              "value": 7
            },
            {
              "name": "rip",
              "pretty_value": "0x555555693755 <LLVMFuzzerTestOneInput+533>",
              "size": 8,
              "type": "",
              "value": 93824993539925
            },
            {
              "name": "eflags",
              "pretty_value": "[ IF RF ]",
              "size": 4,
              "type": "i386_eflags",
              "value": 66050
            },
            {
              "name": "cs",
              "pretty_value": "51",
              "size": 4,
              "type": "int32_t",
              "value": 51
            },
            {
              "name": "ss",
              "pretty_value": "43",
              "size": 4,
              "type": "int32_t",
              "value": 43
            },
            {
              "name": "ds",
              "pretty_value": "0",
              "size": 4,
              "type": "int32_t",
              "value": 0
            },
            {
              "name": "es",
              "pretty_value": "0",
              "size": 4,
              "type": "int32_t",
              "value": 0
            },
            {
              "name": "fs",
              "pretty_value": "0",
              "size": 4,
              "type": "int32_t",
              "value": 0
            },
            {
              "name": "gs",
              "pretty_value": "0",
              "size": 4,
              "type": "int32_t",
              "value": 0
            }
          ],
          "tid": 1
        },
        "stop_info": {
          "faulting_address": 93824993539925,
          "signal_code": 2,
          "signal_name": "SIGILL",
          "signal_number": 4
        }
      },
      "result": "SUCCESS"
    }
  },
  "report_options": {
    "child_output_lines": 25,
    "show_child_output": false
  },
  "testcase": "./test_assets/crashes/crash-1036e40820c11936e0b8d3069623cbecad6b6b95"
}
--- RAWJSON REPORT END ---
[+] Triage stats [Crashes: 1 (unique 1), No crash: 0, Timeout: 0, Errored: 0]`

// LineByLineReader produces a writer that will write exactly one line per read call.
// It is made to sample unbuffered producers.
type LineByLineReader struct {
	data []byte
}

func NewLineByLineReader(s string) *LineByLineReader {
	return &LineByLineReader{
		data: []byte(s),
	}
}
func (l *LineByLineReader) Read(dst []byte) (int, error) {
	if len(l.data) == 0 {
		return 0, io.EOF
	}

	idx := bytes.IndexByte(l.data, '\n')

	n := 0
	if idx < 0 {
		n = copy(dst, l.data)
	} else {
		n = copy(dst, l.data[:idx+1])
	}

	l.data = l.data[n:]

	return n, nil
}

var reportLineCount int

func init() {
	reportLineCount = strings.Count(report, "\n") + 1
}
