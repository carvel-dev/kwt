## pfctl

To reset rules in case kwt stops abruptly:

```
sudo pfctl -sa
sudo pfctl -a kwt-tcp-xxx -F all
```
