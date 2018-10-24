#!/bin/bash

# $OUT_FILE=$1

# build tiff from each file
for FILE in tmp/*.csv; do
    GEOTIFF="${FILE%.*}.tif"
    XYZ="${FILE%.*}.xyz"
    $(echo head -n 1 $FILE) >  "$XYZ"; \
        tail -n +2 $FILE | sort -n -t ',' -k2 -k1 >> "$XYZ"; \
        gdal_translate "$XYZ" "$GEOTIFF";
done

# gdalwarp --config GDAL_CACHEMAX 3000 -wm 3000 tmp/*.tif $OUT_FILE
gdalwarp --config GDAL_CACHEMAX 3000 -wm 3000 tmp/*.tif out.tif
