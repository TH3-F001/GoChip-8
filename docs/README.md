# GoChip8
A Chip-8 interpreter written in go
- Tobias 'The Beast' Langhoff: https://tobiasvl.github.io/blog/write-a-chip-8-emulator/
- Praise to the CowGod!: http://devernay.free.fr/hacks/chip8/C8TECH10.HTM

# Components
- Memory: 4KB
- Display: 64 x 32 (128 x 64 for SUPER-CHIP)
- Program Counter (PC): 16-bit points to current instruction (pointer)
- Index Register (IR): 16-bit register that points to memory locations
- Stack: Holds 16-bit memory addresses used to call and return from subroutines/functions
- Delay Timer (DT): 8-bit timer - decremented at 60hz until it reaches 0
- Sound Timer (ST): 8-bit timer that gives off a beeping sound as long as it isnt 0
- 16 Variable Registers: 8-bit general-purpose registers numbered 0-F (V0-VF)
    - VF is also used as a flag register

---

## Memory
- 4096 Bytes of memory
- all 16 bit pointers and stack entries can only address 12-bits. this is becuase the memory cant exceed 4096 Bytes
- All memory is writable
- Typically the Chip-8 interpreter was stred form memory address 000 to 1FF, and the Chip-8 Program was expected to be loaded in after it starting at memory address 200
    - Best practice is to leave this initial space empty except for the font
- 

## Display
- W64px x H32px
- Each Pixel is either on or off (or a single bit)
    - eg. 256 bytes of vRAM
- Display is redrawn at 60 Hz (60FPS)
    - Some people just redraw the display when the emulator executes an instruction that modifies display data
- DXYN is used to draw sprites on the screen, where each bit corresponds to a hrizontal pixel
- Sprites are betwwen 1 and 15 bytes tall
- Sprites are drawn treeting the screen as all 0 bits, and then flipping to 1 in all the locations of the screen the sprite should be drawn to (xor could prove dodgy side effects, but i bet xor is what they had in mind when they built the thing.)
    - Ammendment, just black out the screen before each redraw. yummy yummy flicker.
    - Flicker could be mitigated via a fade out, or some other solution (the plan is to give the display a phosphorescent look anyway)


### Font
    - The emulator should have a built-in font with sprite data
    - Possible characters are just hex: 0-F
    - Each font character should be 4 pixels wide and 5 tall
    - Font data should be stored in memory
    - Games draw characters like sprites, so they set IR to the character's memory location, and then draw it.
    - There is an instruction for setting IR to a character's address so you can choose where to put it
    - Font is stored somewhere on the first 512 bytes (000-1FF) but it's common to put it in 050-09F
    - Fonts courtesy of **[zZeck] (https://github.com/zZeck)** on github https://github.com/mattmikolay/chip-8/issues/3

## Stack
- FILO data structure with pop and push
- Stack is used to call and return from subroutines
- This is where addresses are saved (16-bit addresses with 12 bits accessible)
- Early interpreters would reserve some memory in the memory space for the stack, but that isnt really needed, the stack can be it's own memory area external from main memory
- The traditional stack was limited to 16 two-byte entries (you dont have to implement this, but doing so prevents buffer overflows)

## Timers
- Both the sound timer and the delay timer work the same way
- timers are one byte in size, and are decremented by one 60 times a second
- Sound timer makes computer beep as long as its above 0
- the interpreter doesnt program the delay. the game just chooses how to use it

## Keypad
1 2 3 C  ->  1 2 3 4

4 5 6 D  ->  Q W E R

7 8 9 E  ->  A S D F

A 0 B F  ->  Z X C V 

- Use scan codes rather than key string constants

## Fetch, Decode, Execute Loop
An emulator's main task is to run an inifitie loop and perform three tasks in succession:
- Fetch instruction from memory at current PC
    - Read the instruction that PC is currently pointing to
    - Instructions are 2 bytes, so you need to combine them into a single 16-bit instruction
    - After reading  the instruction PC shouldbe incremented by 2 so it's ready to fetch the next opcode
- Decode the instruction to find out what the emulator should do.
    - The decode step is more advanced in other systems, but pretty simple in Chip-8
    - Chip-8 isntructions are divided into two categories based on the first nibble (half byte or first hex number)
    - the decode section is basically just a big if-else / switch statement in which different ops are performed based on the first hex number
        - "Mask off (with a “binary AND”) the first number in the instruction, and have one case per number."
        - Even though the first nibble tells you what kind of instruction it is, the rest of the nibbles will have different meanings. To denote these meanings, we usually call the remaining nibbles different things:
    - X: The second nibble. used to look up one of 16 variable registers (VX (from V0 to VF))
    - Y: The third nibble. also used to look up one of the 16 variable registers (VX (from V0 to VF))
    - N: The fourth nibble, a 4-bit number
    - NN: The second byte (Third and fourth nibbles), an 8 bit number
    - NNN: The second, third and fourth nibbles, a 12-bit memory address
    - Note:
        - X and Y are always used to look up values in registers.  (eg: X refers to "memory[VX]" not the value X)
    - Tobias recommends extracting each nibble first, before entering the switch statement to make things easy
- Execute the instruction and do what it says
    - This doesnt have to be a separate step if you did a switch statement in the Decode step
    - Note for other emulators:
        - emulators for other systems might have a bunch of instructions of the same type: for example an op that adds two numbers together might be able to take literal numbers or memory addresses that point to numbers, or a combination (called addressing modes)
        


### Timing
- Original processors ran at 1MHz
- 90s calculator versions ran at 4MHz
- Best to make the speed of the processor configurable
- 700 instructions per second fira well enough for most chip-8 programs
- Probably better to go off of instructions per second rather than MHz