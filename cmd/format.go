package cmd

import (
	"github.com/co-pilot-cli/co-pilot/pkg/maven"
	"github.com/co-pilot-cli/co-pilot/pkg/spring"
	"github.com/spf13/cobra"
)

var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Format functionality for a project",
	Long:  `Format functionality for a project`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := EnableDebug(cmd); err != nil {
			log.Fatalln(err)
		}
		if err := ctx.FindAndPopulateMavenProjects(); err != nil {
			log.Fatalln(err)
		}
	},
}

var formatPomCmd = &cobra.Command{
	Use:   "pom",
	Short: "Formats pom.xml and sorts dependencies",
	Long:  `Formats pom.xml and sorts dependencies`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx.OnEachProject("formatting", nil)
	},
}

var formatVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Removes version tags and replaces them with property tags",
	Long:  `Removes version tags and replaces them with property tags`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx.OnEachProject("removes version tags", maven.ChangeVersionToPropertyTags())
	},
}

var formatInheritVersion = &cobra.Command{
	Use:   "inherit",
	Short: "Removes manual versions from spring dependencies",
	Long:  `Removes manual versions from spring dependencies`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx.OnEachProject("removes manual version from spring dependency", spring.CleanManualVersions())
	},
}

func init() {
	RootCmd.AddCommand(formatCmd)
	formatCmd.AddCommand(formatPomCmd)
	formatCmd.AddCommand(formatVersionCmd)
	formatCmd.AddCommand(formatInheritVersion)

	formatCmd.PersistentFlags().BoolVarP(&ctx.Recursive, "recursive", "r", false, "turn on recursive mode")
	formatCmd.PersistentFlags().StringVar(&ctx.TargetDirectory, "target", ".", "Optional target directory")
	formatCmd.PersistentFlags().BoolVar(&ctx.DryRun, "dry-run", false, "dry run does not write to pom.xml")
}
