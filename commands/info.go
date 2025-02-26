package commands                                                                                   
                                                                                                    
 import (                                                                                           
     "fmt"                                                                                          
     "io/ioutil"                                                                                    
     "strings"                                                                                      
                                                                                                    
     "github.com/spf13/cobra"                                                                       
 )                                                                                                  
                                                                                                    
 // InfoCmd represents the info command                                                             
 var InfoCmd = &cobra.Command{                                                                      
     Use:   "info",                                                                                 
     Short: "Displays information about projects and available scripts",                            
     RunE: func(cmd *cobra.Command, args []string) error {                                          
         return displayInfo()                                                                       
     },                                                                                             
 }                                                                                                  
                                                                                                    
 // displayInfo lists the number of projects and available scripts                                  
 func displayInfo() error {                                                                         
     projects, err := loadProjects()                                                                
     if err != nil {                                                                                
         return fmt.Errorf("failed to load projects: %w", err)                                      
     }                                                                                              
                                                                                                    
     fmt.Printf("Number of projects: %d\n", len(projects.Projects))                                 
                                                                                                    
     files, err := ioutil.ReadDir(scriptsFolder)                                                    
     if err != nil {                                                                                
         return fmt.Errorf("failed to read scripts directory: %w", err)                             
     }                                                                                              
                                                                                                    
     fmt.Println("Available scripts:")                                                              
     for _, file := range files {                                                                   
         if !file.IsDir() && strings.HasSuffix(file.Name(), ".ts") {                                
             fmt.Printf("- %s\n", file.Name())                                                      
         }                                                                                          
     }                                                                                              
                                                                                                    
     return nil                                                                                     
 }     