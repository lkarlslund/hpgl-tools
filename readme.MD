# HPGL Tools

*This repository is some small tools I needed while playing around with my vintage Roland DXY-980 pen plotter. After 35 years, this thing is still going strong, which is quite impressive.*

### My current workflow is as follows:
- Import vector graphics into Inkscape 1.2
- If I need multi color prints, I assign each color to a dedicated layer named "Pen 1 Speed 100" which indicates to the HPGL exporter what pen and speed to use
- I then choose File, Save As... and HPGL as the output format
- In the export options window, I have everything under "Plot features" set to 0, except for curve flatness which I've set to 0.5
- Then I optimize the HPGL file, in order to speed up the print (see below)
- Finally I send the HPGL file to the plotter (see below)

So you'll need Inkscape and this repository in order to replicate this

## OPTIMIZER

The optimizer takes an input file, and produces an output file. You can choose whether the optimizer is allowed to break consecutive lines apart, in order to optimize better, and also whether lines are allowed to be drawn in reverse. Both options are enabled default, as this is a pen plotter and it makes no difference to the output. If you're using this for a cutting plotter, you should probably disable both due to the knife offset coming into play.

```
hpgl-optimizer --input my-unoptimized-inkscape.hpgl --output awesome-optimized.hpgl [--reversible=true] [--breakandassemble=true]
```

## SENDER

As I had a problem finding a simple utility that just sends a file over serial with hardware handshake support and no other fuss, I ended up writing my own. It defaults to the settings I need, but feel free to play around with this. It does *not* support Xon/Xoff software handshake, as I didn't need it, but PRs are welcome.

To list available ports, just run it without options. To send a file:

```
hpgl-sender --input awesome-optimized.hpgl [-] --port COM5 [--baud 9600] [--stopbits 1] [--databits 8] [--parity N] [--dtr true] [--rts true] [--dsr true] [--cts true] [--maxchunk 16]
```

DTS and RTS options set the bits on the sending side, and if DSR or CTS is true the program waits for them to be high before sending data, implementing hardware handshake. It sends up to 'maxchunk' bytes at a time, which defaults to 16 bytes.

## Have fun!

