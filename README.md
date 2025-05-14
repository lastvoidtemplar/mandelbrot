```bash
go run main.go mandelbrot.go queue.go image.go -iter=500 -st-x=-0.3 -st-y=-1 -end-x=0.1 -end-y=-0.6 -result-width=8000 -result-height=8000 -gran-width=50 -gran-height=50 -export -threads=14 -distribute=dump -topology=hypercube
```
```bash
go run main.go mandelbrot.go queue.go image.go -iter=500 -st-x=-0.05 -st-y=-0.825 -end-x=0 -end-y=-0.775 -result-width=8000 -result-height=8000 -gran-width=50 -gran-height=50 -export -threads=12 -distribute=dump -topology=hypercube
```


