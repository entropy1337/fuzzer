// Copyright 2015 syzkaller project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	flagLinux = flag.String("linux", "", "path to linux kernel checkout")
)

func main() {
	flag.Parse()
	if *flagLinux == "" {
		failf("provide path to linux kernel checkout via -linux flag (or make generate LINUX= flag)")
	}
	if len(flag.Args()) == 0 {
		failf("usage: sysgen -linux=linux_checkout input_file")
	}

	var r io.Reader
	for i, f := range flag.Args() {
		inf, err := os.Open(f)
		if err != nil {
			failf("failed to open input file: %v", err)
		}
		defer inf.Close()
		if i == 0 {
			r = bufio.NewReader(inf)
		} else {
			r = io.MultiReader(r, bufio.NewReader(inf))
		}
	}

	includes, defines, syscalls, structs, unnamed, flags := parse(r)
	intFlags, flagVals := compileFlags(includes, defines, flags)

	out := new(bytes.Buffer)
	generate(syscalls, structs, unnamed, intFlags, flagVals, out)
	writeSource("sys/sys.go", out.Bytes())

	out = new(bytes.Buffer)
	generateConsts(flagVals, out)
	writeSource("prog/consts.go", out.Bytes())

	generateSyscallsNumbers(syscalls)
}

type Syscall struct {
	Name     string
	CallName string
	Args     [][]string
	Ret      []string
}

type Struct struct {
	Name    string
	Flds    [][]string
	IsUnion bool
	Packed  bool
	Varlen  bool
	Align   int
}

func generate(syscalls []Syscall, structs map[string]Struct, unnamed map[string][]string, flags map[string][]string, flagVals map[string]string, out io.Writer) {
	fmt.Fprintf(out, "// AUTOGENERATED FILE\n")
	fmt.Fprintf(out, "package sys\n\n")

	fmt.Fprintf(out, "var Calls []*Call\n")
	fmt.Fprintf(out, "func initCalls() {\n")
	for i, s := range syscalls {
		fmt.Fprintf(out, "func() { Calls = append(Calls, &Call{ID: %v, Name: \"%v\", CallName: \"%v\"", i, s.Name, s.CallName)
		if len(s.Ret) != 0 {
			fmt.Fprintf(out, ", Ret: ")
			generateArg("ret", s.Ret[0], s.Ret[1:], structs, unnamed, flags, flagVals, false, out)
		}
		fmt.Fprintf(out, ", Args: []Type{")
		for i, a := range s.Args {
			if i != 0 {
				fmt.Fprintf(out, ", ")
			}
			generateArg(a[0], a[1], a[2:], structs, unnamed, flags, flagVals, false, out)
		}
		fmt.Fprintf(out, "}})}()\n")
	}
	fmt.Fprintf(out, "}\n")
}

func generateArg(name, typ string, a []string, structs map[string]Struct, unnamed map[string][]string, flags map[string][]string, flagVals map[string]string, isField bool, out io.Writer) {
	name = "\"" + name + "\""
	opt := false
	for i, v := range a {
		if v == "opt" {
			opt = true
			copy(a[i:], a[i+1:])
			a = a[:len(a)-1]
			break
		}
	}
	common := func() string {
		return fmt.Sprintf("TypeCommon: TypeCommon{TypeName: %v, IsOptional: %v}", name, opt)
	}
	switch typ {
	case "fd":
		if len(a) == 0 {
			a = append(a, "")
		}
		if want := 1; len(a) != want {
			failf("wrong number of arguments for %v arg %v want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "ResourceType{%v, Kind: ResFD, Subkind: %v}", common(), fmtFdKind(a[0]))
	case "io_ctx":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "ResourceType{%v, Kind: ResIOCtx}", common())
	case "ipc":
		if want := 1; len(a) != want {
			failf("wrong number of arguments for %v arg %v want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "ResourceType{%v, Kind: ResIPC, Subkind: %v}", common(), fmtIPCKind(a[0]))
	case "key":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "ResourceType{%v, Kind: ResKey}", common())
	case "inotifydesc":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "ResourceType{%v, Kind: ResInotifyDesc}", common())
	case "timerid":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "ResourceType{%v, Kind: ResTimerid}", common())
	case "pid":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "ResourceType{%v, Kind: ResPid}", common())
	case "uid":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "ResourceType{%v, Kind: ResUid}", common())
	case "gid":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "ResourceType{%v, Kind: ResGid}", common())
	case "drmctx":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "ResourceType{%v, Kind: ResDrmCtx}", common())
	case "fileoff":
		var size uint64
		if isField {
			if want := 2; len(a) != want {
				failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
			}
			size = typeToSize(a[1])
		} else {
			if want := 1; len(a) != want {
				failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
			}
		}
		fmt.Fprintf(out, "FileoffType{%v, File: \"%v\", TypeSize: %v}", common(), a[0], size)
	case "buffer":
		if want := 1; len(a) != want {
			failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
		}
		commonHdr := common()
		opt = false
		fmt.Fprintf(out, "PtrType{%v, Dir: %v, Type: BufferType{%v, Kind: BufferBlob}}", commonHdr, fmtDir(a[0]), common())
	case "string":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
		}
		commonHdr := common()
		opt = false
		fmt.Fprintf(out, "PtrType{%v, Dir: %v, Type: BufferType{%v, Kind: BufferString}}", commonHdr, fmtDir("in"), common())
	case "filesystem":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
		}
		commonHdr := common()
		opt = false
		fmt.Fprintf(out, "PtrType{%v, Dir: %v, Type: BufferType{%v, Kind: BufferFilesystem}}", commonHdr, fmtDir("in"), common())
	case "sockaddr":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "BufferType{%v, Kind: BufferSockaddr}", common())
	case "salg_type":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "BufferType{%v, Kind: BufferAlgType}", common())
	case "salg_name":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "BufferType{%v, Kind: BufferAlgName}", common())
	case "vma":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "VmaType{%v}", common())
	case "len", "bytesize":
		var size uint64
		if isField {
			if want := 2; len(a) != want {
				failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
			}
			size = typeToSize(a[1])
		} else {
			if want := 1; len(a) != want {
				failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
			}
		}
		fmt.Fprintf(out, "LenType{%v, Buf: \"%v\", TypeSize: %v, ByteSize: %v}", common(), a[0], size, typ == "bytesize")
	case "flags":
		var size uint64
		if isField {
			if want := 2; len(a) != want {
				failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
			}
			size = typeToSize(a[1])
		} else {
			if want := 1; len(a) != want {
				failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
			}
		}
		vals := flags[a[0]]
		if len(vals) == 0 {
			failf("unknown flag %v", a[0])
		}
		fmt.Fprintf(out, "FlagsType{%v, TypeSize: %v, Vals: []uintptr{%v}}", common(), size, strings.Join(vals, ","))
	case "const":
		var size uint64
		if isField {
			if want := 2; len(a) != want {
				failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
			}
			size = typeToSize(a[1])
		} else {
			if want := 1; len(a) != want {
				failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
			}
		}
		val := flagVals[a[0]]
		if val == "" {
			val = a[0]
		}
		fmt.Fprintf(out, "ConstType{%v, TypeSize: %v, Val: uintptr(%v)}", common(), size, val)
	case "strconst":
		if want := 1; len(a) != want {
			failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "PtrType{%v, Dir: %v, Type: StrConstType{%v, Val: \"%v\"}}", common(), fmtDir("in"), common(), a[0]+"\\x00")
	case "int8", "int16", "int32", "int64", "intptr":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "IntType{%v, TypeSize: %v}", common(), typeToSize(typ))
	case "signalno":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "IntType{%v, TypeSize: 4, Kind: IntSignalno}", common())
	case "in_addr":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "IntType{%v, TypeSize: 4, Kind: IntInaddr}", common())
	case "in_port":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "IntType{%v, TypeSize: 2, Kind: IntInport}", common())
	case "filename":
		if want := 0; len(a) != want {
			failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
		}
		commonHdr := common()
		opt = false
		fmt.Fprintf(out, "PtrType{%v, Dir: DirIn, Type: FilenameType{%v}}", commonHdr, common())
	case "array":
		want := 1
		if len(a) == 2 {
			want = 2
		}
		if len(a) != want {
			failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
		}
		sz := "0"
		if len(a) == 2 {
			sz = flagVals[a[1]]
			if sz == "" {
				sz = a[1]
			}
		}
		fmt.Fprintf(out, "ArrayType{%v, Type: %v, Len: %v}", common(), generateType(a[0], structs, unnamed, flags, flagVals), sz)
	case "ptr":
		if want := 2; len(a) != want {
			failf("wrong number of arguments for %v arg %v, want %v, got %v", typ, name, want, len(a))
		}
		fmt.Fprintf(out, "PtrType{%v, Type: %v, Dir: %v}", common(), generateType(a[1], structs, unnamed, flags, flagVals), fmtDir(a[0]))
	default:
		if strings.HasPrefix(typ, "unnamed") {
			if inner, ok := unnamed[typ]; ok {
				generateArg("", inner[0], inner[1:], structs, unnamed, flags, flagVals, isField, out)
				return
			}
			failf("unknown unnamed type '%v'", typ)
		}
		if str, ok := structs[typ]; ok {
			typ := "StructType"
			fields := "Fields"
			if str.IsUnion {
				typ = "UnionType"
				fields = "Options"
			}
			packed := ""
			if str.Packed {
				packed = ", packed: true"
			}
			varlen := ""
			if str.Varlen {
				varlen = ", varlen: true"
			}
			align := ""
			if str.Align != 0 {
				align = fmt.Sprintf(", align: %v", str.Align)
			}
			fmt.Fprintf(out, "%v{TypeCommon: TypeCommon{TypeName: \"%v\", IsOptional: %v} %v %v %v, %v: []Type{", typ, str.Name, false, packed, align, varlen, fields)
			for i, a := range str.Flds {
				if i != 0 {
					fmt.Fprintf(out, ", ")
				}
				generateArg(a[0], a[1], a[2:], structs, unnamed, flags, flagVals, true, out)
			}
			fmt.Fprintf(out, "}}")
			return
		}
		failf("unknown arg type \"%v\" for %v", typ, name)
	}
}

func generateType(typ string, structs map[string]Struct, unnamed map[string][]string, flags map[string][]string, flagVals map[string]string) string {
	buf := new(bytes.Buffer)
	generateArg("", typ, nil, structs, unnamed, flags, flagVals, true, buf)
	return buf.String()
}

func fmtFdKind(s string) string {
	switch s {
	case "":
		return "ResAny"
	case "file":
		return "FdFile"
	case "sock":
		return "FdSock"
	case "pipe":
		return "FdPipe"
	case "signal":
		return "FdSignal"
	case "event":
		return "FdEvent"
	case "timer":
		return "FdTimer"
	case "epoll":
		return "FdEpoll"
	case "dir":
		return "FdDir"
	case "mq":
		return "FdMq"
	case "inotify":
		return "FdInotify"
	case "fanotify":
		return "FdFanotify"
	case "tty":
		return "FdTty"
	case "dri":
		return "FdDRI"
	case "fuse":
		return "FdFuse"
	case "kdbus":
		return "FdKdbus"
	case "bpf_map":
		return "FdBpfMap"
	case "bpf_prog":
		return "FdBpfProg"
	case "perf":
		return "FdPerf"
	case "uffd":
		return "FdUserFault"
	case "alg":
		return "FdAlg"
	case "algconn":
		return "FdAlgConn"
	case "nfc_raw":
		return "FdNfcRaw"
	case "nfc_llcp":
		return "FdNfcLlcp"
	case "bt_hci":
		return "FdBtHci"
	case "bt_sco":
		return "FdBtSco"
	case "bt_l2cap":
		return "FdBtL2cap"
	case "bt_rfcomm":
		return "FdBtRfcomm"
	case "bt_hidp":
		return "FdBtHidp"
	case "bt_cmtp":
		return "FdBtCmtp"
	case "bt_bnep":
		return "FdBtBnep"
	case "unix":
		return "FdUnix"
	case "sctp":
		return "FdSctp"
	case "netlink":
		return "FdNetlink"
	case "kvm":
		return "FdKvm"
	case "kvmvm":
		return "FdKvmVm"
	case "kvmcpu":
		return "FdKvmCpu"
	case "sndseq":
		return "FdSndSeq"
	case "sndtimer":
		return "FdSndTimer"
	case "sndctrl":
		return "FdSndControl"
	case "evdev":
		return "FdInputEvent"
	case "tun":
		return "FdTun"
	case "random":
		return "FdRandom"
	default:
		failf("bad fd type %v", s)
		return ""
	}
}

func fmtIPCKind(s string) string {
	switch s {
	case "msq":
		return "IPCMsq"
	case "sem":
		return "IPCSem"
	case "shm":
		return "IPCShm"
	default:
		failf("bad ipc type %v", s)
		return ""
	}
}

func fmtDir(s string) string {
	switch s {
	case "in":
		return "DirIn"
	case "out":
		return "DirOut"
	case "inout":
		return "DirInOut"
	default:
		failf("bad direction %v", s)
		return ""
	}
}

func typeToSize(typ string) uint64 {
	switch typ {
	case "int8", "int16", "int32", "int64", "intptr":
	default:
		failf("unknown type %v", typ)
	}
	sz := int64(64) // TODO: assume that pointer is 8 bytes for now
	if typ != "intptr" {
		sz, _ = strconv.ParseInt(typ[3:], 10, 64)
	}
	return uint64(sz / 8)
}

type F struct {
	name string
	val  string
}

type FlagArray []F

func (a FlagArray) Len() int           { return len(a) }
func (a FlagArray) Less(i, j int) bool { return a[i].name < a[j].name }
func (a FlagArray) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type SortedSyscall struct {
	name string
	nr   int
}

func generateConsts(flags map[string]string, out io.Writer) {
	var ff []F
	for k, v := range flags {
		ff = append(ff, F{k, v})
	}
	sort.Sort(FlagArray(ff))

	fmt.Fprintf(out, "// AUTOGENERATED FILE\n")
	fmt.Fprintf(out, "package prog\n\n")
	fmt.Fprintf(out, "const (\n")
	for _, f := range ff {
		fmt.Fprintf(out, "	%v = %v\n", f.name, f.val)
	}
	fmt.Fprintf(out, ")\n")
	fmt.Fprintf(out, "\n")
}

func compileFlags(includes []string, defines map[string]string, flags map[string][]string) (map[string][]string, map[string]string) {
	vals := make(map[string]string)
	for _, fvals := range flags {
		for _, v := range fvals {
			vals[v] = ""
		}
	}
	for k := range defines {
		vals[k] = ""
	}
	valArray := make([]string, 0, len(vals))
	for k := range vals {
		valArray = append(valArray, k)
	}
	// TODO: should use target arch
	flagVals := fetchValues("x86", valArray, includes, defines)
	for i, f := range valArray {
		vals[f] = flagVals[i]
	}
	res := make(map[string][]string)
	for fname, fvals := range flags {
		var arr []string
		for _, v := range fvals {
			arr = append(arr, vals[v])
		}
		if res[fname] != nil {
			failf("flag %v is defined multiple times", fname)
		}
		res[fname] = arr
	}
	ids := make(map[string]string)
	for k, v := range vals {
		if isIdentifier(k) {
			ids[k] = v
		}
	}
	return res, ids
}

func isIdentifier(s string) bool {
	for i, c := range s {
		if c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || i > 0 && (c >= '0' && c <= '9') {
			continue
		}
		return false
	}
	return true
}

func parse(in io.Reader) (includes []string, defines map[string]string, syscalls []Syscall, structs map[string]Struct, unnamed map[string][]string, flags map[string][]string) {
	p := NewParser(in)
	defines = make(map[string]string)
	structs = make(map[string]Struct)
	unnamed = make(map[string][]string)
	flags = make(map[string][]string)
	var str *Struct
	for p.Scan() {
		if p.EOF() || p.Char() == '#' {
			continue
		}
		if str != nil {
			// Parsing a struct.
			if p.Char() == '}' || p.Char() == ']' {
				p.Parse(p.Char())
				for _, attr := range parseType1(p, unnamed, flags, "")[1:] {
					if str.IsUnion {
						switch attr {
						case "varlen":
							str.Varlen = true
						default:
							failf("unknown union %v attribute: %v", str.Name, attr)
						}
					} else {
						switch attr {
						case "packed":
							str.Packed = true
						case "align_1":
							str.Align = 1
						case "align_2":
							str.Align = 2
						case "align_4":
							str.Align = 4
						case "align_8":
							str.Align = 8
						default:
							failf("unknown struct %v attribute: %v", str.Name, attr)
						}
					}
				}
				structs[str.Name] = *str
				str = nil
			} else {
				p.SkipWs()
				fld := []string{p.Ident()}
				fld = append(fld, parseType(p, unnamed, flags)...)
				str.Flds = append(str.Flds, fld)
			}
		} else {
			name := p.Ident()
			if name == "include" {
				p.Parse('<')
				var include []byte
				for {
					ch := p.Char()
					if ch == '>' {
						break
					}
					p.Parse(ch)
					include = append(include, ch)
				}
				p.Parse('>')
				includes = append(includes, string(include))
			} else if name == "define" {
				key := p.Ident()
				var val []byte
				for !p.EOF() {
					ch := p.Char()
					p.Parse(ch)
					val = append(val, ch)
				}
				if defines[key] != "" {
					failf("%v define is defined multiple times", key)
				}
				defines[key] = fmt.Sprintf("(%s)", val)
			} else {
				switch ch := p.Char(); ch {
				case '(':
					// syscall
					p.Parse('(')
					var args [][]string
					for p.Char() != ')' {
						arg := []string{p.Ident()}
						arg = append(arg, parseType(p, unnamed, flags)...)
						args = append(args, arg)
						if p.Char() != ')' {
							p.Parse(',')
						}
					}
					p.Parse(')')
					var ret []string
					if !p.EOF() {
						ret = parseType(p, unnamed, flags)
					}
					callName := name
					if idx := strings.IndexByte(callName, '$'); idx != -1 {
						callName = callName[:idx]
					}
					syscalls = append(syscalls, Syscall{name, callName, args, ret})
				case '=':
					// flag
					p.Parse('=')
					vals := []string{p.Ident()}
					for !p.EOF() {
						p.Parse(',')
						vals = append(vals, p.Ident())
					}
					flags[name] = vals
				case '{', '[':
					p.Parse(ch)
					if _, ok := structs[name]; ok {
						failf("%v struct is defined multiple times", name)
					}
					str = &Struct{Name: name, IsUnion: ch == '['}
				default:
					failf("bad line (%v)", p.Str())
				}
			}
		}
		if !p.EOF() {
			failf("trailing data (%v)", p.Str())
		}
	}
	return
}

func parseType(p *Parser, unnamed map[string][]string, flags map[string][]string) []string {
	return parseType1(p, unnamed, flags, p.Ident())
}

var (
	unnamedSeq int
	constSeq   int
)

func parseType1(p *Parser, unnamed map[string][]string, flags map[string][]string, name string) []string {
	typ := []string{name}
	if !p.EOF() && p.Char() == '[' {
		p.Parse('[')
		for {
			id := p.Ident()
			if p.Char() == '[' {
				inner := parseType1(p, unnamed, flags, id)
				id = fmt.Sprintf("unnamed%v", unnamedSeq)
				unnamedSeq++
				unnamed[id] = inner
			}
			typ = append(typ, id)
			if p.Char() == ']' {
				break
			}
			p.Parse(',')
		}
		p.Parse(']')
	}
	if name == "const" && len(typ) > 1 {
		// Create a fake flag with the const value.
		id := fmt.Sprintf("const_flag_%v", constSeq)
		constSeq++
		flags[id] = typ[1:2]
	}
	if name == "array" && len(typ) > 2 {
		// Create a fake flag with the const value.
		id := fmt.Sprintf("const_flag_%v", constSeq)
		constSeq++
		flags[id] = typ[2:3]
	}
	return typ
}

func writeSource(file string, data []byte) {
	src, err := format.Source(data)
	if err != nil {
		fmt.Printf("%s\n", data)
		failf("failed to format output: %v", err)
	}
	writeFile(file, src)
}

func writeFile(file string, data []byte) {
	outf, err := os.Create(file)
	if err != nil {
		failf("failed to create output file: %v", err)
	}
	defer outf.Close()
	outf.Write(data)
}

func failf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}