FROM scratch

COPY outfitpicker /outfitpicker
COPY outfitpicker-admin /outfitpicker-admin

ENTRYPOINT ["/outfitpicker"]