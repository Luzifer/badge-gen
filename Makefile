documentation:
	echo "| Key | Type | Description |" >config.md
	echo "| --- | ---- | ----------- |" >>config.md
	awk -F ' - ' '/#configStore/{ gsub(/.*configStore /, "", $$1); printf "| %s | %s | %s |\n", $$1, $$2, $$3 }' *.go | sort >>config.md
