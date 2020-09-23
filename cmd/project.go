package cmd

import (
	"co-pilot/pkg/config"
	"co-pilot/pkg/logger"
	"co-pilot/pkg/maven"
	"fmt"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project [OPTIONS]",
	Short: "Project options",
	Long:  `Various project helper commands`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := EnableDebug(cmd); err != nil {
			log.Fatalln(err)
		}
		ctx.FindAndPopulatePomProjects()
	},
}

var projectInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a maven project with co-pilot files and formatting",
	Long:  `Initializes a maven project with co-pilot files and formatting`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, project := range ctx.Projects {
			log.Info(logger.White(fmt.Sprintf("formating pom file %s", project.PomFile)))

			if !ctx.DryRun {
				projectCfg := config.InitProjectConfigurationFromModel(project.PomModel)

				if err := projectCfg.WriteTo(project.ConfigFile); err != nil {
					log.Warnln(err)
					continue
				}

				if err := maven.Write(project, ctx.Overwrite); err != nil {
					log.Warnln(err)
				}
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectInitCmd)

	projectCmd.PersistentFlags().BoolVarP(&ctx.Recursive, "recursive", "r", false, "turn on recursive mode")
	projectCmd.PersistentFlags().StringVar(&ctx.TargetDirectory, "target", ".", "Optional target directory")
	projectCmd.PersistentFlags().BoolVar(&ctx.Overwrite, "overwrite", true, "Overwrite pom.xml file")
	projectCmd.PersistentFlags().BoolVar(&ctx.DryRun, "dry-run", false, "dry run does not write to pom.xml")
}
