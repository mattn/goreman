#!perl

my $app = sub {
  [200, [], ["hello world $AUTHOR"]]
}
