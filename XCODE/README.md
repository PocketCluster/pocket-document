# Xcode Tips

**When CocoaPod goes whacky, we might want to reset everything and start over**

1. Deintegrated cocoa pods, using `pod deintegrate`. Check this link <https://github.com/kylef/cocoapods-deintegrate>  
2. Search on the target settings and project file for `pod`. Remove anything that looked like it belonged to cocoa pods.  
3. `pod install` once again.
  
> Reference  

- <http://stackoverflow.com/questions/32768516/cocoapods-ld-library-not-found-for-lpods-objectivesugar>