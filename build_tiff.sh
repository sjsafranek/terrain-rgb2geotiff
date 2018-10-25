#!/bin/bash

OUT_FILE=$1

# build tiff from each file
echo "Building tif files from csv map tiles"
for FILE in tmp/*.csv; do
    GEOTIFF="${FILE%.*}.tif"
    XYZ="${FILE%.*}.xyz"

    echo "Building $XYZ from $FILE"
    $(echo head -n 1 $FILE) >  "$XYZ"; \
        tail -n +2 $FILE | sort -n -t ',' -k2 -k1 >> "$XYZ";

    echo "Building $GEOTIFF from $XYZ"
    gdal_translate "$XYZ" "$GEOTIFF"
done

echo "Merging tif files to $OUT_FILE"
gdalwarp --config GDAL_CACHEMAX 3000 -wm 3000 tmp/*.tif $OUT_FILE

# echo "Clean up"
rm tmp/*.csv
rm tmp/*.xyz
rm tmp/*.tif
