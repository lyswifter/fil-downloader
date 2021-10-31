This is the tool to download files from qiniu cruster manually.

toCheck = []string{
    sealPath,
    filepath.Join(cachePath, "p_aux"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-0.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-1.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-2.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-3.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-4.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-5.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-6.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-7.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-8.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-9.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-10.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-11.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-12.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-13.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-14.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-15.dat"),
}

toCheck = []string{
    sealPath,
    filepath.Join(cachePath, "p_aux"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-0.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-1.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-2.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-3.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-4.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-5.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-6.dat"),
    filepath.Join(cachePath, "sc-02-data-tree-r-last-7.dat"),
}

//http://10.10.36.69:5000/getfile/12/f0111007//cache/s-t0111007-85654/p_aux
//http://10.10.36.76:5000/getfile/12/f0111007//cache/s-t0111007-2638/sc-02-data-tree-r-last-3.dat
//http://10.10.24.95:5000/getfile/12/f0111007//sealed/s-t0111007-2640