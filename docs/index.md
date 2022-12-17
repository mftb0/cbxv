# cbxv user manual

## tl;dr - Quickstart
Open the cbxv archive (.tar.gz, .zip, .dmg), usually by dobule-clicking on it

Locate the executable (cbxv, cbxv.exe, cbxv.app), double-click on it

cbxv will start and you'll see:

<img align="center" width="496" src="cbxv_ss_02.png">

Hit the "o" key or click the "File Button" in the upper-right

The File Open Dialog will display, navigate to a cbx (.cbr, .cbz) file, double-click it

You'll see something like:

<img align="center" width="496" src="cbxv_ss_03.png">
<sub>Note: All comic images shown on these pages are believed to be in the public
domain. If you feel that's in error, please notify me and I will replace them
with ones that are.</sub>

In this case we're seeing the cover on the left and the inside cover on the right. 
Hit the "r" key (Join Toggle || Button) and you'll see:

<img align="center" width="496" src="cbxv_ss_04.png">

The first page, the cover, has essentially been "Joined" or turned into a single-page
with a span of 2, so that it displays by itself. All of the other pages have also been
adjusted in the layout. Hit the "Right Arrow" key and you'll see:

<img align="center" width="496" src="cbxv_ss_05.png">

The inside front cover is now on the left and the first page on the right, just 
as it should be. If the book you're reading has other pages that are out of 
place you can join them or hide them as necessary until the layout is correct. 
cbxv will remember the layout so the next time you open the book it will be 
correct.

Excelsior!

## Dependencies
- Linux - You must have Gtk3 installed. This is very common on Linux. If you 
    don't have it already you must install the appropriate package for your 
    distro.

    - Arch and Fedora   - gtk3
    - Debian and Ubuntu - libgtk-3-0

- Windows and Mac - Everything that you need is in the archive available under 
    the releases section.

## Installation
-   Linux - Download the Linux build from the release area and unarchive it. 
    On Linux cbxv is a single executable, put it wherever you like and run it. 
    For your convenience a simple script is provided to put a desktop file and 
    icon in the appropriate places for your user.

-   Windows - Download the Windows build from the release area and unarchive it. 
    Copy the resulting directory to program files or wherever you like and run it.

-   Mac - Download the Mac build from the release area and unarchive it. Copy the 
    cbxv.app directory to Applications and double-click on it.

## Interface Elements

<img align="center" width="496" src="prg_elements-03.png">

## Commands

### File Commands
- openFile            
    The openFile command when given no arguments will prompt you with the 
    fileOpen dialog to provide a file path. After a path is provided any file
    open in the UI will be closed and the new file will be opened and loaded
    into the interface. 

    The fileOpen dialog by default is configured to restrict choosing only files 
    that end with an appropriate extension, but you can change it to all files. 
    Whatever file you specify either with the fileDialog or from the command 
    line cbxv will try to open it. If it's a valid file with an inappropriate 
    extension like .zip it may very well succeed. If it's simply an invalid file 
    it will fail.

    Keys: o  
    Mouse: The File Button  
    CLI: If you start cbxv from the command line you can provide a path and it 
        will be opened.  

- closeFile           
    The closeFile command will close any open file and unload it

    Keys: c

### Navigation Commands
    cbxv is a viewer, most of what you do is navigating around so you can read
    the comic you have loaded. Consequently there are quite a few keys dedicated
    to basic navigation. 3 "sets" in fact; "Standard Keys" - Arrow Cluster, 
    "Gamer Keys" - "wasd", and "Vi Keys" - "hjkl".

- rightPage           
    Always takes you one page to the right. If you have the reading Direction
    set to Left-To-Right, then it will take you to the next page. If you toggle
    the reading Direction to Right-To-Left, then going a page to the right will
    take you to the previous page.  

    This key can also trigger next file or previous file when at the end or
    beggining of the comic.

    Keys: Right Arrow or d or l

- leftPage  
    Always takes you one page to the left. If you have the reading Direction
    set to Left-To-Right, then it will take you to the previous page. If you 
    toggle the reading Direction to Right-To-Left, then going a page to the 
    left will take you to the next page.

    This key can also trigger next file or previous file when at the end or
    beggining of the comic.

    Keys: Left Arrow or w or h

- firstPage  
    Always takes you to the first page

    Keys: Up Arrow or w or k

- lastPage  
    Always takes you to the last page

    Keys: Up Arrow or w or k

- nextFile  
    Whenever you open a cbx file cbxv creates a sorted list of all the cbx files
    in the same directory and the position of the current file in that list. The
    nextFile command takes you to the next cbx file in the list.

    Keys: n

- previousFile  
    Whenever you open a cbx file cbxv creates a sorted list of all the cbx files
    in the same directory and the position of the current file in that list. The
    previousFile command takes you to the previous cbx file in the list.

    Keys: p

### Page Commands
- selectPage          [Tab]               NA
- exportPage          e                   NA

### Bookmark Commands
- toggleBookmark      [Space]             Bookmark Buttons
- lastBookmark        L                   NA

### Layout Commands
- Direction           [BackTick]          Direction Button
- 1-Page Layout       1                   NA
- 2-Page Layout       2                   NA
- stripLayout         3                   NA
- hidePage            -                   NA
- toggleJoin          r                   Join Toggle

### General Commands
- quit                q                   Window Close Button 
- help                ?|[F1]              Question Mark Button
- toggleFullscreen    f                   Fullscreen Toggle

