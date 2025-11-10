if [ -z "$(ls -A ./internal/generated/user_settings)" ]; then
	echo "missing generated protobuf packages"
	exit 1
fi
