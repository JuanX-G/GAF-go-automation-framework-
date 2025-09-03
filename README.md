# GAF-go-automation-framework
A simple and minimal automation framework written in go that relies on go for scripting.

### Example uses 
    - fetching urls with news e. g. "RSS"
    - running cleanup
    - showing the weather, time and calendar etc.
## Required folder structure
    - root (any name)
        |
        |- cmd
            |- main.go
        |- internal
            | - [prog_modules]
        |- pkgs
        |- vendor 
        |- configs
            |- actions/ #here your declare your own function to be used in scripts. File names must be gaa_[your_name].go
            |- automations/ #here you define your full scripts using the smaller function in 'actions'. File names must follow gaf_[your_name].txt
            |- declaration/ #here you declare the shape of your self-defined functions. File names must be 'gadf.txt'
            |- main.conf
