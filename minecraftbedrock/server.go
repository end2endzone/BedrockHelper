package minecraftbedrock

type Server struct {
	Path string
}

func (s Server) IsValid() bool {
	valid := IsValidServerDirectory(s.Path)
	return valid
}

func (s Server) ActiveWorld() (*World, error) {
	worldDir, err := FindActiveWorldDir(s.Path)
	if err != nil {
		return nil, err
	}

	world := &World{
		Path: worldDir,
	}

	return world, nil
}

func GetServer(path string) (*Server, error) {
	err := ValidateServerDirectory(path)
	if err != nil {
		return nil, err
	}

	server := &Server{
		Path: path,
	}
	return server, nil
}
