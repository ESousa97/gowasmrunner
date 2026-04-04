package main

import (
	"log"
	"os"
)

// This generator produces a WebAssembly binary that implements an iterative
// Fibonacci function. The module represents what a Rust toolchain (rustc
// --target wasm32-unknown-unknown) would emit for:
//
//   #[no_mangle]
//   pub extern "C" fn fibonacci(n: u32) -> u32 {
//       if n <= 1 { return n; }
//       let (mut a, mut b) = (0u32, 1u32);
//       for _ in 2..=n {
//           let tmp = a + b;
//           a = b;
//           b = tmp;
//       }
//       b
//   }
//
// WAT representation:
//
//   (module
//     (func $fibonacci (param $n i32) (result i32)
//       (local $a i32)
//       (local $b i32)
//       (local $i i32)
//
//       ;; Base case: if n <= 1, return n
//       (if (i32.le_u (local.get $n) (i32.const 1))
//         (then (return (local.get $n)))
//       )
//
//       ;; Initialize: a = 0, b = 1, i = 2
//       (local.set $a (i32.const 0))
//       (local.set $b (i32.const 1))
//       (local.set $i (i32.const 2))
//
//       ;; Loop: while i <= n
//       (block $break
//         (loop $continue
//           (br_if $break (i32.gt_u (local.get $i) (local.get $n)))
//           (local.set $a (local.get $b))         ;; save old b in a
//           ... (compute new b = old_a + old_b)
//           (local.set $i (i32.add (local.get $i) (i32.const 1)))
//           (br $continue)
//         )
//       )
//
//       (local.get $b)
//     )
//     (export "fibonacci" (func $fibonacci))
//   )

func main() {
	// Build the code body first, then compute sizes accurately.
	// Param: $n (index 0), Locals: $a (1), $b (2), $i (3)

	code := []byte{
		// --- local declarations ---
		0x03,       // 3 local declaration groups
		0x01, 0x7f, // 1x i32 ($a = local 1)
		0x01, 0x7f, // 1x i32 ($b = local 2)
		0x01, 0x7f, // 1x i32 ($i = local 3)

		// --- if (n <= 1) return n ---
		0x20, 0x00, // local.get 0 ($n)
		0x41, 0x01, // i32.const 1
		0x4d,       // i32.le_u
		0x04, 0x40, // if (void block type)
		0x20, 0x00, // local.get 0 ($n)
		0x0f, // return
		0x0b, // end if

		// --- a = 0 ---
		0x41, 0x00, // i32.const 0
		0x21, 0x01, // local.set 1 ($a)

		// --- b = 1 ---
		0x41, 0x01, // i32.const 1
		0x21, 0x02, // local.set 2 ($b)

		// --- i = 2 ---
		0x41, 0x02, // i32.const 2
		0x21, 0x03, // local.set 3 ($i)

		// --- block $break ---
		0x02, 0x40, // block (void)

		// --- loop $continue ---
		0x03, 0x40, // loop (void)

		// br_if $break: if (i > n)
		0x20, 0x03, // local.get 3 ($i)
		0x20, 0x00, // local.get 0 ($n)
		0x4b,       // i32.gt_u
		0x0d, 0x01, // br_if 1 (to $break)

		// tmp = a + b  (we use the stack to avoid needing a 4th local)
		// Push: b, then compute a+b, set b=a+b, set a=old_b
		0x20, 0x02, // local.get 2 ($b)  -- save old b on stack
		0x20, 0x01, // local.get 1 ($a)
		0x20, 0x02, // local.get 2 ($b)
		0x6a,       // i32.add           -- stack: old_b, (a+b)
		0x21, 0x02, // local.set 2 ($b)  -- b = a+b
		0x21, 0x01, // local.set 1 ($a)  -- a = old_b

		// i = i + 1
		0x20, 0x03, // local.get 3 ($i)
		0x41, 0x01, // i32.const 1
		0x6a,       // i32.add
		0x21, 0x03, // local.set 3 ($i)

		// br $continue
		0x0c, 0x00, // br 0 (to $continue)

		0x0b, // end loop
		0x0b, // end block

		// return b
		0x20, 0x02, // local.get 2 ($b)
		0x0b, // end function
	}

	bodySize := byte(len(code))

	// Export name "fibonacci"
	exportName := []byte{0x66, 0x69, 0x62, 0x6f, 0x6e, 0x61, 0x63, 0x63, 0x69} // "fibonacci"

	// Build full module
	var wasm []byte

	// Magic + Version
	wasm = append(wasm, 0x00, 0x61, 0x73, 0x6d) // \0asm
	wasm = append(wasm, 0x01, 0x00, 0x00, 0x00) // version 1

	// Type Section: 1 type (i32) -> (i32)
	wasm = append(wasm,
		0x01,       // section id: type
		0x06,       // section size
		0x01,       // 1 type
		0x60,       // func type
		0x01, 0x7f, // 1 param: i32
		0x01, 0x7f, // 1 result: i32
	)

	// Function Section: 1 function, type index 0
	wasm = append(wasm,
		0x03, // section id: function
		0x02, // section size
		0x01, // 1 function
		0x00, // type index 0
	)

	// Export Section: "fibonacci" -> func 0
	exportSectionPayload := []byte{
		0x01,                  // 1 export
		byte(len(exportName)), // name length
	}
	exportSectionPayload = append(exportSectionPayload, exportName...)
	exportSectionPayload = append(exportSectionPayload, 0x00) // kind: func
	exportSectionPayload = append(exportSectionPayload, 0x00) // index: 0

	wasm = append(wasm, 0x07)                            // section id: export
	wasm = append(wasm, byte(len(exportSectionPayload))) // section size
	wasm = append(wasm, exportSectionPayload...)

	// Code Section: 1 body
	codeSectionPayload := []byte{
		0x01,     // 1 function body
		bodySize, // body size
	}
	codeSectionPayload = append(codeSectionPayload, code...)

	wasm = append(wasm, 0x0a)                          // section id: code
	wasm = append(wasm, byte(len(codeSectionPayload))) // section size
	wasm = append(wasm, codeSectionPayload...)

	if err := os.WriteFile("examples/rust_fibonacci.wasm", wasm, 0644); err != nil {
		log.Fatalf("failed to write wasm file: %v", err)
	}
	log.Println("Created examples/rust_fibonacci.wasm (Rust-style Fibonacci)")
}
