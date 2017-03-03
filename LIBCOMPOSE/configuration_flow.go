ctx.Context{
    Context:project.Context{
        ComposeFiles: []string{"pocket-deploy.yml"}, 
    }
}
docker.NewProject(context *ctx.Context, parseOptions *config.ParseOptions) (project.APIProject, error)
    project.NewProject(context *Context, runtime RuntimeProject, parseOptions *config.ParseOptions) (*project.Project)
    project.Parse()
        project.Context.open()
            project.Context.readComposeFiles()
            project.Context.determineProject()

        project.load(file string, bytes []byte) error
            config.Merge(existingServices *ServiceConfigs, environmentLookup EnvironmentLookup, resourceLookup ResourceLookup, file string, bytes []byte, options *ParseOptions) 
                (string, map[string]*ServiceConfig, map[string]*VolumeConfig, map[string]*NetworkConfig, error)

                config.MergeServicesV2(existingServices *ServiceConfigs, environmentLookup EnvironmentLookup, resourceLookup ResourceLookup, file string, datas RawServiceMap, options *ParseOptions) 
                    (map[string]*ServiceConfig, error)

                    config.parseV2(resourceLookup ResourceLookup, environmentLookup EnvironmentLookup, inFile string, serviceData RawService, datas RawServiceMap, options *ParseOptions) 
                        (RawService, error)

                        config.resolveContextV2(inFile string, serviceData RawService) (RawService)

                config.MergeServicesV1(existingServices *ServiceConfigs, environmentLookup EnvironmentLookup, resourceLookup ResourceLookup, file string, datas RawServiceMap, options *ParseOptions) 
                    (map[string]*ServiceConfigV1, error)

    context.LookupConfig()

err = project.Up(context.Background(), options.Up{})