	PROCESSOR 6524T8

	; dasm24 Apple2four.asm -oApple2four.rom -f3

KBD         = $c000		; R last key pressed + 128
KBDSTRB     = $c010		; RW keyboard strobe
TAPEOUT     = $c020		; RW toggle caseette tape output
SPKR        = $c030		; RW toggle speaker
TXTCLR      = $c050		; RW display graphics
TXTSET      = $c051		; RW display text
MIXSET      = $c053		; RW display split screen
TXTPAGE1    = $c054		; RW display page 1
LORES       = $c056		; RW display lo-res graphics
SETAN0      = $c058		; RW annunciator 0 off
SETAN1      = $c05a		; RW annunciator 1 off
CLRAN2      = $c05d		; RW annunciator 2 on
CLRAN3      = $c05f		; RW annunciator 3 on
TAPEIN      = $c060		; R cassete tape input
PADDL0      = $c064		; R analog input 0
PTRIG       = $c070		; RW analog input reset
CLRROM      = $cfff		; disable slot C8 ROM
RSTVECTOR   = $fffc  	; Apple ][ 6502 reset vector
PWREDUP     = $03f4  	; Apple ][ stores #$A5 to make RST a soft reboot

	SEG Code
	ORG $FF0000

Start
	jmp Reset
	jmp ClearScreen
	jmp ScrollScreen

	ORG $FF0200

Reset
	cld
	ldx #$ff
	txs
	lda TXTSET
	lda TXTPAGE1
	lda SETAN0					; AN0 = TTL hi
	lda SETAN1					; AN1 = TTL hi
	lda CLRAN2					; AN2 = TTL lo
	lda CLRAN3					; AN3 = TTL lo
	lda CLRROM					; turn off extension ROM
	bit KBDSTRB		; clear keyboard
	ldx #0
	jsr ClearScreen	; clear the screen (x = row 0)

AppleII4
	ldx #0
.loop_AppleII4
	lda Hello,x
	beq .done_AppleII4
	ora #$80
	sta $400,x
	inx
	bne .loop_AppleII4
.done_AppleII4

EchoKeys
	lda #2
	sta $02						; $02 = TEXT ROW (start at row 2)
	lda #0
	sta $03						; $03 = TEXT COLUMN (start at column 0)
	jsr ClearInputBuffer
.echo_loop
	ldx $02						; $02 = TEXT ROW
	lda TextScreenBaseL,x
	sta $00						; $00/$01 = TEXT line base address
	lda TextScreenBaseH,x
	sta $01
	lda $03
	bne .echo_read
.draw_prompt
	ldy #0
	lda #$BA 					; ':'
	sta ($00),y
	iny
	sty $03
.echo_read
	bit KBD 					; check keyboard for next key
	bpl .echo_read
	lda KBD 					; get the key code from the keyboard
	bit KBDSTRB					; clear keyboard strobe (a.k.a. ack keyboard read)
	cmp #$8D					; 13 = CR ($80 | $0D in high ASCII)
	beq .echo_next_line
	cmp #$88					; 8 = BS ($80 | $08 in high ASCII)
	beq .echo_backspace
	cmp #$9B					; 27 = ESC ($80 | $1B in high ASCII)
	beq .echo_escape
	cmp #$99					; control character ($80 | $19 in high ASCII)
	bcc .echo_read
	ldy $03
	cpy #40 					; ignore if >= column 40
	beq .echo_loop
	sta ($00),y
	and #$7F					; store (low) ASCII in TEXT buffer
	dey
	sta $200,y
	inc $03						; next column
	bra .echo_loop
.echo_next_line
	jsr NextLine
.echo_command
	jsr CommandLine
	jsr ClearInputBuffer
	jmp .echo_loop
.echo_backspace
	lda $03
	beq .echo_loop
	tay
	dey
	sty $03
	lda #$A0
	sta ($00),y
	jmp .echo_loop
.echo_escape
	ldx #2
	jsr ClearScreen				; clear the screen (x = row 2)
	lda #2
	sta $02						; reset to row 2
	lda #0
	sta $03						; reset to column 0
	jmp .echo_loop

CommandLine
.command_check_clear
	lda #<CMD_Clear				; $00/$01 = COMMAND string base address
	sta $05
	lda #>CMD_Clear
	sta $06
	lda #$FF 					; TODO: Add >> to dasm
	sta $07
	jsr CompareText;
	beq .not_command_clear
	jsr DoClear
	rts
.not_command_clear
.command_check_memory
	lda #<CMD_Memory
	sta $05
	lda #>CMD_Memory
	sta $06
	lda #$FF 					; TODO: Add >> to dasm
	sta $07
	jsr CompareText;
	beq .not_command_memory
	jsr DoMemory
	rts
.not_command_memory
.command_check_reset
	lda #<CMD_Reset
	sta $05
	lda #>CMD_Reset
	sta $06
	lda #$FF 					; TODO: Add >> to dasm
	sta $07
	jsr CompareText;
	beq .not_command_reset
	jsr DoReset
	rts
.not_command_reset
.command_check_2plus
	lda #<CMD_2Plus
	sta $05
	lda #>CMD_2Plus
	sta $06
	lda #$FF 					; TODO: Add >> to dasm
	sta $07
	jsr CompareText;
	beq .not_command_2plus
	jsr Do2Plus
	rts
.not_command_2plus
	rts

CompareText
	ldy #0
.compare_loop
	lda $200,y
 	a24
	cmp ($05),y
	bne .compare_no_match
	iny
	cmp #0
	bne .compare_loop
.compare_match
	lda ($05),y
	bne .compare_no_match
	lda #$FF
	rts
.compare_no_match
	lda #0
	rts

DoClear
	ldx #2
	jsr ClearScreen
	lda #2
	sta $02						; $02 = TEXT ROW (start at row 2)
	lda #0
	sta $03						; $03 = TEXT COLUMN (start at column 0)
	rts

DoMemory
	jsr NextLine
	lda #0
	jsr PrintHexByte
	lda #$AD
	jsr PrintChar
	jsr PrintSpace
	ldx #0
.loop_Memory
	lda $00,x
	jsr PrintHexByte
	jsr PrintSpace
	inx
	cpx #8
	bne .loop_Memory
	jsr NextLine
	rts

PrintHexByte					; A = byte // $00/$01 = address of the line on the TEXT screen // $03 = TEXT COLUMN
	pha
	lsr
	lsr
	lsr
	lsr
	jsr PrintHexDigit
	pla
	and #$0F
	jsr PrintHexDigit
	rts

PrintHexDigit					; A = nibble // $00/$01 = address of the line on the TEXT screen // $03 = TEXT COLUMN
	cmp #10
	bcc .print_hex_digit_09
	clc
	adc #$B7
	bra .print_hex_digit
.print_hex_digit_09
	adc #$B0
.print_hex_digit
	ldy $03
	sta ($00),y
	inc $03
	rts

PrintChar						; A = ASCII // $00/$01 = address of the line on the TEXT screen // $03 = TEXT COLUMN
	ldy $03
	sta ($00),y
	inc $03
	rts

PrintSpace						; $00/$01 = address of the line on the TEXT screen // $03 = TEXT COLUMN
	lda #$A0
	ldy $03
	sta ($00),y
	inc $03
	rts

DoReset
  lda #$A1
  sta $425
  sta $426
  sta $427
	jmp Reset

Do2Plus
	sws							; Reset the stack width to 8-bits
	ldx #$ff 					; Reset the stack to $1FF
	txs
	lda #0
	sta PWREDUP 				; Make sure the Apple ][ ROM thinks this is a fresh reboot
	jmp (RSTVECTOR)				; Jump to the 6502 64K RST vector

ClearScreen						; X = row // Y = column // $00/$01 = address of the line on the TEXT screen
.loop_clear_line
	lda TextScreenBaseL,x
	sta $00
	lda TextScreenBaseH,x
	sta $01
	lda #$A0 		; $20 (space) | $80 (high ASCII)
	ldy #0
.loop_clear_char
	sta ($00),y
	iny
	cpy #40
	bne .loop_clear_char
	inx
	cpx #24
	bne .loop_clear_line
	rts

NextLine 						; $00/$01 = address of the line on the TEXT screen // $02 = TEXT ROW // $03 = TEXT COLUMN
	lda $02
	cmp #23
	beq .next_line_scroll
.next_line_next_row
	inc $02						; increment the current row
	lda #0
	sta $03						; reset to column 0
	ldx $02
	lda TextScreenBaseL,x
	sta $00
	lda TextScreenBaseH,x
	sta $01
	rts
.next_line_scroll
	ldx #2
	jsr ScrollScreen			; scroll the screen (x = row 2)
	lda #23
	sta $02						; reset to row 23
	lda #00
	sta $03						; reset to column 0
	rts

ScrollScreen					; X = first row // Y = column
.loop_scroll_line
	lda TextScreenBaseL,x 		; $00/$01 = address of the line on the TEXT screen
	sta $00
	lda TextScreenBaseH,x
	sta $01
	lda TextScreenBaseL+1,x
	sta $04						; $04/$05 = address of the next on the TEXT screen
	lda TextScreenBaseH+1,x
	sta $05
	ldy #0
.loop_scroll_char
	lda ($04),y
	sta ($00),y
	iny
	cpy #40
	bne .loop_scroll_char
	inx
	cpx #23
	bne .loop_scroll_line
.scroll_clear_last_line
	lda TextScreenBaseL,X 		; $00/$01 = address of the line on the TEXT screen
	sta $00
	lda TextScreenBaseH,x
	sta $01
	lda #$20
	ldy #0
.loop_scroll_clear_line
	sta ($00),y
	iny
	cpy #40
	bne .loop_scroll_clear_line
.done_with_scroll
	rts

ClearInputBuffer
	lda #0
	ldx #$ff
.loop_clear_buffer
	sta $200,x					; $200 = TEXT BUFFER ($2FF = length / $2nn = characters)
	dex
	bne .loop_clear_buffer
	sta $200
	rts


	SEG Data
	ORG $FF8000

TextScreenBaseL
	DC.B $00, $80, $00, $80, $00, $80, $00, $80
	DC.B $28, $A8, $28, $A8, $28, $A8, $28, $A8
	DC.B $50, $D0, $50, $D0, $50, $D0, $50, $D0
TextScreenBaseH
	DC.B $04, $04, $05, $05, $06, $06, $07, $07
	DC.B $04, $04, $05, $05, $06, $06, $07, $07
	DC.B $04, $04, $05, $05, $06, $06, $07, $07

Hello DC.B "Apple ][4", $00

CommandListLength
	DC.B 4
CommandList
	DC.W CMD_Clear
	DC.W CMD_Memory
	DC.W CMD_Reset
	DC.W CMD_2Plus

CMD_Clear DC.B "clear", $00
CMD_Memory DC.B "memory", $00
CMD_Reset DC.B "reset", $00
CMD_2Plus DC.B "2+", $00

	SEG Interrupts
	ORG $FFFFF7

Vectors
	.byte $00, $00, $00 		; NMI
	.byte $00, $00, $FF 		; RESET
	.byte $00, $00, $00 		; IRQ

	END